package linearAlgebra

import "math"

type Vec3 [3]float64

type Point3 = Vec3
type Color = Vec3

func NewVec3(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

func (v Vec3) X() float64 {
	return v[0]
}

func (v Vec3) Y() float64 {
	return v[1]
}

func (v Vec3) Z() float64 {
	return v[2]
}

func (v Vec3) Add(o Vec3) Vec3 {
	return Vec3{
		v[0] + o[0],
		v[1] + o[1],
		v[2] + o[2],
	}
}

func (v Vec3) Sub(o Vec3) Vec3 {
	return Vec3{
		v[0] - o[0],
		v[1] - o[1],
		v[2] - o[2],
	}
}

func (v Vec3) MultScalar(s float64) Vec3 {
	return Vec3{
		v[0] * s,
		v[1] * s,
		v[2] * s,
	}
}

func (v Vec3) DivScalar(s float64) Vec3 {
	return Vec3{
		v[0] / s,
		v[1] / s,
		v[2] / s,
	}
}

func (v Vec3) Dot(o Vec3) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		v[1]*o[2] - v[2]*o[1],
		v[2]*o[0] - v[0]*o[2],
		v[0]*o[1] - v[1]*o[0],
	}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.Dot(v))
}

func (v Vec3) LengthSquared() float64 {
	return v.Dot(v)
}

func (v Vec3) Normalize() Vec3 {
	return v.DivScalar(v.Length())
}
