package sample

type Fuel struct {
	Type   string
	Amount float32
}

type Rocket struct {
	Power int
	Name  string
	Direction struct {
		X float32
		Y float32
		Z float32
	}
	Tank Fuel
	V    []int
}

type Control interface {
	Land()
	IsLanded() (success bool)
	Aircraft() (*Rocket)

	Launch(rocket *Rocket) (bool, error)
}
