package main

type gufFile struct {
	Deps []dep
}

type dep struct {
	Local   string
	Remote  string
	Version string
}
