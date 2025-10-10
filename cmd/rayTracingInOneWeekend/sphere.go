package main

import (
	"math"

	"github.com/BadgerBadgerBadgerBadger/goplay/internal/linearAlgebra"
)

// Recall that the equation for a sphere centered at the origin of radius 𝑅
// is 𝑥2+𝑦2+𝑧2=𝑅2. Put another way, if a given point (𝑥,𝑦,𝑧) is on the sphere,
// then 𝑥2+𝑦2+𝑧2=𝑅2 . If the given point (𝑥,𝑦,𝑧) is inside the sphere, then 𝑥2+𝑦2+𝑧2<𝑅2,
// and if a given point (𝑥,𝑦,𝑧) is outside the sphere, then 𝑥2+𝑦2+𝑧2>𝑅2.
// It gets uglier if the sphere center is at (𝐶𝑥,𝐶𝑦,𝐶𝑧) :
// (𝑥−𝐶𝑥)2+(𝑦−𝐶𝑦)2+(𝑧−𝐶𝑧)2=𝑟2
//
// In graphics, you almost always want your formulas to be in terms of vectors so
// all the x/y/z stuff is under the hood in the vec3 class. You might note that
// the vector from center 𝐂=(𝐶𝑥,𝐶𝑦,𝐶𝑧) to point 𝐏=(𝑥,𝑦,𝑧) is (𝐏−𝐂), and therefore
// (𝐏−𝐂)⋅(𝐏−𝐂)=(𝑥−𝐶𝑥)2+(𝑦−𝐶𝑦)2+(𝑧−𝐶𝑧)2
// So the equation of the sphere in vector form is:
// (𝐏−𝐂)⋅(𝐏−𝐂)=𝑟2
// We can read this as “any point 𝐏
// that satisfies this equation is on the sphere”. We want to know if our ray 𝐏(𝑡)=𝐀+𝑡𝐛
// ever hits the sphere anywhere. If it does hit the sphere, there is some 𝑡
// for which 𝐏(𝑡) satisfies the sphere equation. So we are looking for any 𝑡
// where this is true:
// (𝐏(𝑡)−𝐂)⋅(𝐏(𝑡)−𝐂)=𝑟2
// or expanding the full form of the ray 𝐏(𝑡) :
// (𝐀+𝑡𝐛−𝐂)⋅(𝐀+𝑡𝐛−𝐂)=𝑟2
// The rules of vector algebra are all that we would want here. If we expand that equation and
// move all the terms to the left hand side we get:
// 𝑡2𝐛⋅𝐛+2𝑡𝐛⋅(𝐀−𝐂)+(𝐀−𝐂)⋅(𝐀−𝐂)−𝑟2=0
// The vectors and 𝑟
// in that equation are all constant and known. The unknown is 𝑡
// , and the equation is a quadratic, like you probably saw in your high school math class. You can
// solve for 𝑡 and there is a square root part that is either positive (meaning two real solutions),
// negative (meaning no real solutions), or zero (meaning one real solution).
func hitSphere(center linearAlgebra.Vec3, radius float64, ray Ray) float64 {
	// (ray.origin - center) * (ray.origin - center) = radius * radius
	oc := ray.Origin.Sub(center)
	// a = ray.dir * ray.dir
	a := ray.Dir.LengthSquared()
	// b = 2 * oc * ray.dir
	b := 2 * oc.Dot(ray.Dir)
	// c = oc * oc - radius * radius
	c := oc.LengthSquared() - radius*radius
	// b*b - 4*a*c
	discriminant := b*b - 4*a*c

	if discriminant < 0 {
		return -1.0
	} else {
		return (-b - math.Sqrt(discriminant)) / (2.0 * a)
	}
}
