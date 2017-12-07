//
// Some package description
//
package sample

const Greeting = "HEllo!" // Greeting value

// BBBBBBBBBBB
type Fuel struct {
	Type   string
	Amount float32 //AAAAAAAAA
}

// Rocket - This is a ROCKET!
//adsasd
//asdasd
type Rocket struct {
	Power int // This is power
	Name  string
	// inline
	Direction struct {
		X float32
		Y float32
		Z float32
	}
	Tank Fuel
	V    []int
	D    map[int]string
}

// Control?
type Control interface {
	//AA;;
	Land()
	IsLanded() (success bool)
	Aircraft() (*Rocket)

	Launch(rocket *Rocket) (bool, error)
}
