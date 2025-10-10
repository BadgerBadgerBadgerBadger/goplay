package main

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BadgerBadgerBadgerBadger/goplay/pkg/util"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/cobra"
)

// TOTPEntry represents a stored TOTP secret entry with issuer and account info
type TOTPEntry struct {
	Issuer      string `json:"issuer"`                 // Service name (e.g., GitHub, Google)
	AccountName string `json:"account_name"`           // Account identifier (e.g., email, username)
	ServiceName string `json:"service_name,omitempty"` // Internal name used for keychain storage
}

// Name of the keychain item that stores the registry of all TOTP entries
const registryServiceName = "totp-cli-registry"

// main is the entry point of the application
func main() {
	ctx := context.Background()
	err := run(ctx, os.Args, os.Stdout)
	util.ExitOnError(err, "failed to run program")
}

// run sets up the command line interface and executes the selected command
// It configures all commands, flags, and autocompletion functionality
func run(ctx context.Context, args []string, out io.Writer) error {
	var rootCmd = &cobra.Command{
		Use:   "totp-cli",
		Short: "TOTP CLI generates time-based one-time passwords",
		Long: `A command-line tool for managing TOTP secrets and generating 
time-based one-time passwords. Secrets are stored securely in the macOS Keychain.`,
	}

	// Add command
	var issuer, account, secret string
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new TOTP entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTOTP(issuer, account, secret, out)
		},
	}
	addCmd.Flags().StringVarP(&issuer, "issuer", "i", "", "Service name (e.g. GitHub)")
	addCmd.Flags().StringVarP(&account, "account", "a", "", "Account identifier (e.g. username or email)")
	addCmd.Flags().StringVarP(&secret, "secret", "s", "", "TOTP secret key")
	addCmd.MarkFlagRequired("issuer")
	addCmd.MarkFlagRequired("account")
	addCmd.MarkFlagRequired("secret")

	// Generate command
	var genIssuer, genAccount string
	var listCodes bool
	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate TOTP code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTOTPCode(genIssuer, genAccount, listCodes, out)
		},
	}
	genCmd.Flags().StringVarP(&genIssuer, "issuer", "i", "", "Service name to generate code for")
	genCmd.Flags().StringVarP(&genAccount, "account", "a", "", "Account to generate code for")
	genCmd.Flags().BoolVarP(&listCodes, "list", "l", false, "List multiple codes with their valid times")
	genCmd.MarkFlagRequired("issuer")
	genCmd.MarkFlagRequired("account")

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved TOTP entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTOTPs(out)
		},
	}

	// Remove command
	var rmIssuer, rmAccount string
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a TOTP entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeTOTP(rmIssuer, rmAccount, out)
		},
	}
	removeCmd.Flags().StringVarP(&rmIssuer, "issuer", "i", "", "Service name to remove")
	removeCmd.Flags().StringVarP(&rmAccount, "account", "a", "", "Account to remove")
	removeCmd.MarkFlagRequired("issuer")
	removeCmd.MarkFlagRequired("account")

	// Completion command
	var shellType string
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for bash or zsh.

To load completions:

Bash:
  $ source <(totp-cli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ totp-cli completion bash > /etc/bash_completion.d/totp-cli
  # macOS:
  $ totp-cli completion bash > $(brew --prefix)/etc/bash_completion.d/totp-cli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ totp-cli completion zsh > "${fpath[1]}/_totp-cli"

  # You will need to start a new shell for this setup to take effect.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch shellType {
			case "bash":
				return rootCmd.GenBashCompletion(out)
			case "zsh":
				return rootCmd.GenZshCompletion(out)
			default:
				return errors.Errorf("unsupported shell type: %s (supported: bash, zsh)", shellType)
			}
		},
	}
	completionCmd.Flags().StringVarP(&shellType, "shell", "s", "bash", "Shell type (bash or zsh)")

	// Add commands to root
	rootCmd.AddCommand(addCmd, genCmd, listCmd, removeCmd, completionCmd)

	// Enable command autocompletion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add custom autocompletion for issuer and account flags
	_ = genCmd.RegisterFlagCompletionFunc("issuer", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		issuers, err := getIssuers()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return issuers, cobra.ShellCompDirectiveNoFileComp
	})

	_ = genCmd.RegisterFlagCompletionFunc("account", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		issuerFlag := cmd.Flag("issuer")
		if issuerFlag == nil || issuerFlag.Value.String() == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		accounts, err := getAccountsForIssuer(issuerFlag.Value.String())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return accounts, cobra.ShellCompDirectiveNoFileComp
	})

	// Also add completion for remove command
	_ = removeCmd.RegisterFlagCompletionFunc("issuer", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		issuers, err := getIssuers()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return issuers, cobra.ShellCompDirectiveNoFileComp
	})

	_ = removeCmd.RegisterFlagCompletionFunc("account", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		issuerFlag := cmd.Flag("issuer")
		if issuerFlag == nil || issuerFlag.Value.String() == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		accounts, err := getAccountsForIssuer(issuerFlag.Value.String())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return accounts, cobra.ShellCompDirectiveNoFileComp
	})

	// Execute the root command
	return rootCmd.Execute()
}

// getIssuers returns a list of unique issuer names from the registry
// Used for command autocompletion of the issuer flag
func getIssuers() ([]string, error) {
	entries, err := getTOTPRegistry()
	if err != nil {
		return nil, err
	}

	// Create a map to deduplicate issuers
	issuerMap := make(map[string]bool)
	for _, entry := range entries {
		issuerMap[entry.Issuer] = true
	}

	// Convert map to slice
	issuers := make([]string, 0, len(issuerMap))
	for issuer := range issuerMap {
		issuers = append(issuers, issuer)
	}

	return issuers, nil
}

// getAccountsForIssuer returns a list of account names for a given issuer
// Used for context-aware autocompletion of the account flag
func getAccountsForIssuer(issuer string) ([]string, error) {
	entries, err := getTOTPRegistry()
	if err != nil {
		return nil, err
	}

	var accounts []string
	for _, entry := range entries {
		if entry.Issuer == issuer {
			accounts = append(accounts, entry.AccountName)
		}
	}

	return accounts, nil
}

// addTOTP creates a new TOTP entry and stores it in the registry
// It also saves the secret to the macOS Keychain
func addTOTP(issuer, account, secret string, out io.Writer) error {
	// Clean up secret
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	// Validate secret
	if !validateSecret(secret) {
		return errors.New("invalid TOTP secret key format")
	}

	// Create keychain service name
	serviceName := fmt.Sprintf("totp-cli-%s-%s", issuer, account)
	entry := TOTPEntry{
		Issuer:      issuer,
		AccountName: account,
		ServiceName: serviceName,
	}

	// Store in keychain
	err := addKeychainItem(serviceName, secret, "")
	if err != nil {
		return errors.Wrap(err, "failed to store secret in keychain")
	}

	// Update registry
	err = updateTOTPRegistry(entry)
	if err != nil {
		return errors.Wrap(err, "failed to update registry")
	}

	fmt.Fprintf(out, "Added TOTP for %s (%s)\n", issuer, account)
	return nil
}

// generateTOTPCode generates and outputs a time-based one-time password
// for the specified issuer and account
// When listCodes is true, it shows the current code and the next two
func generateTOTPCode(issuer, account string, listCodes bool, out io.Writer) error {
	// Find entry in registry
	entries, err := getTOTPRegistry()
	if err != nil {
		return errors.Wrap(err, "TOTP registry not found")
	}

	var serviceName string
	found := false
	for _, entry := range entries {
		if entry.Issuer == issuer && entry.AccountName == account {
			serviceName = entry.ServiceName
			found = true
			break
		}
	}

	if !found || serviceName == "" {
		return errors.Errorf("TOTP entry not found for %s (%s)", issuer, account)
	}

	// Get secret from keychain
	secret, err := getKeychainSecret(serviceName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve secret for %s (%s)", issuer, account)
	}

	if listCodes {
		// Show current code and next two
		now := time.Now()
		fmt.Fprintf(out, "TOTP codes for %s (%s):\n", issuer, account)
		for i := 0; i < 3; i++ {
			codeTime := now.Add(time.Duration(i*30) * time.Second)
			validUntil := codeTime.Add(30 * time.Second)

			codeAtTime, err := totp.GenerateCode(secret, codeTime)
			if err != nil {
				return errors.Wrap(err, "failed to generate TOTP code")
			}
			fmt.Fprintf(out, "%s (valid until %s)\n",
				codeAtTime,
				validUntil.Format("15:04:05"))
		}
	} else {
		// Generate single code
		code, err := totp.GenerateCode(secret, time.Now())
		if err != nil {
			return errors.Wrap(err, "failed to generate TOTP code")
		}
		fmt.Fprintln(out, code)
	}
	return nil
}

// listTOTPs outputs all stored TOTP entries to the provided writer
// If no entries are found, it displays an appropriate message
func listTOTPs(out io.Writer) error {
	// Get list of entries in registry
	entriesList, err := getTOTPRegistry()
	if err != nil {
		// Registry doesn't exist yet, create it
		fmt.Fprintln(out, "No TOTP entries found.")
		return nil
	}

	if len(entriesList) == 0 {
		fmt.Fprintln(out, "No TOTP entries found.")
		return nil
	}

	fmt.Fprintln(out, "Available TOTP entries:")
	for _, entry := range entriesList {
		fmt.Fprintf(out, "- %s (%s)\n", entry.Issuer, entry.AccountName)
	}
	return nil
}

// removeTOTP deletes a TOTP entry from both the keychain and the registry
// It returns an error if the entry cannot be removed or is not found
func removeTOTP(issuer, account string, out io.Writer) error {
	entries, err := getTOTPRegistry()
	if err != nil {
		fmt.Fprintln(out, "No TOTP entries found.")
		return nil
	}

	var serviceNameToRemove string
	var newEntries []TOTPEntry

	for _, entry := range entries {
		if entry.Issuer == issuer && entry.AccountName == account {
			serviceNameToRemove = entry.ServiceName
		} else {
			newEntries = append(newEntries, entry)
		}
	}

	if serviceNameToRemove == "" {
		fmt.Fprintf(out, "Entry not found: %s (%s)\n", issuer, account)
		return nil
	}

	// Remove from keychain
	cmd := exec.Command("security", "delete-generic-password",
		"-s", serviceNameToRemove)
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to remove entry from keychain for %s (%s)", issuer, account)
	}

	// Update registry
	entriesJSON, err := json.Marshal(newEntries)
	if err != nil {
		return errors.Wrap(err, "failed to marshal registry data")
	}

	// Delete existing registry
	exec.Command("security", "delete-generic-password", "-s", registryServiceName).Run()

	// Add updated registry
	err = addKeychainItem(registryServiceName, string(entriesJSON), "registry")
	if err != nil {
		return errors.Wrap(err, "failed to update registry after removal")
	}

	fmt.Fprintf(out, "Removed TOTP for %s (%s)\n", issuer, account)
	return nil
}

// addKeychainItem stores a secret in the macOS Keychain
// It creates a generic password item with the given service name and secret
// If account is not provided, it defaults to "totp-cli"
func addKeychainItem(service, secret, account string) error {
	if account == "" {
		account = "totp-cli"
	}

	cmd := exec.Command(
		"security", "add-generic-password",
		"-s", service,
		"-a", account,
		"-w", secret,
		"-U",
	)

	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to store secret in keychain for service %s", service)
	}

	return nil
}

// getKeychainSecret retrieves a secret from the macOS Keychain
// It returns the secret as a string, or an error if the item cannot be found
func getKeychainSecret(service string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", service,
		"-w")

	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to retrieve secret for service %s", service)
	}

	return strings.TrimSpace(string(output)), nil
}

// getTOTPRegistry retrieves the list of all TOTP entries from the registry.
// The registry is stored as a JSON-encoded array in the macOS Keychain
// Returns an empty array and error if the registry doesn't exist or can't be parsed.
func getTOTPRegistry() ([]TOTPEntry, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", registryServiceName,
		"-w")

	output, err := cmd.Output()
	if err != nil {
		return []TOTPEntry{}, errors.Wrap(err, "failed to retrieve TOTP registry from keychain")
	}

	var entries []TOTPEntry
	err = json.Unmarshal(output, &entries)
	if err != nil {
		return []TOTPEntry{}, errors.Wrap(err, "failed to parse TOTP registry data")
	}

	return entries, nil
}

// updateTOTPRegistry adds or updates an entry in the TOTP registry
// If the entry already exists (same issuer and account), it's updated
// Otherwise, a new entry is added to the registry
func updateTOTPRegistry(newEntry TOTPEntry) error {
	entries, err := getTOTPRegistry()
	if err != nil {
		// Registry doesn't exist yet, create it
		entries = []TOTPEntry{}
	}

	// Check if entry already exists
	found := false
	for _, entry := range entries {
		if entry.Issuer == newEntry.Issuer && entry.AccountName == newEntry.AccountName {
			found = true
			break
		}
	}

	// Store updated registry only if something changed
	if !found {
		entries = append(entries, newEntry)

		entriesJSON, err := json.Marshal(entries)
		if err != nil {
			return errors.Wrap(err, "failed to marshal registry data")
		}

		// Delete existing registry and add updated one
		exec.Command("security", "delete-generic-password", "-s", registryServiceName).Run()
		return addKeychainItem(registryServiceName, string(entriesJSON), "registry")
	}

	return nil
}

// validateSecret checks if the provided TOTP secret is a valid base32 string
// It automatically adds padding if needed to make the string valid
// Returns true if the secret is valid, false otherwise
func validateSecret(secret string) bool {
	// Handle base32 padding if needed
	if !strings.HasSuffix(secret, "=") {
		padding := 8 - (len(secret) % 8)
		if padding < 8 {
			secret = secret + strings.Repeat("=", padding)
		}
	}

	// Validate the secret is valid base32
	_, err := base32.StdEncoding.DecodeString(secret)
	return err == nil
}
