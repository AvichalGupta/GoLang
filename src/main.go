package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type User struct {
	ID       string
	Name     string
	Age      int
	Gender   string
	District string
	State    string
}

type VaccinationCenter struct {
	ID       string
	District string
	State    string
}

type Appointment struct {
	CenterID string
	UserID   string
	Day      int
}

type System struct {
	Users              map[string]User
	VaccinationCenters map[string]VaccinationCenter
	Capacity           map[string]map[int]int
	Appointments       map[string][]Appointment
	mu                 sync.Mutex
}

func AddUser(vs *System, id string, name string, gender string, age string, state string, district string) (bool, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.Users[id]; exists {
		return false, errors.New(`User with ID already exists`)
	}

	ageInt, err := strconv.Atoi(age)
	if err != nil {
		return false, errors.New(`Invalid age value`)
	}

	if ageInt <= 18 {
		return false, errors.New(`User's age cannot be under 18`)
	}

	vs.Users[id] = User{
		ID:       id,
		Name:     name,
		Age:      ageInt,
		Gender:   gender,
		District: district,
		State:    state,
	}

	fmt.Println("Users Data: ", vs.Users)

	return true, nil
}

func AddVC(vs *System, state string, district string, id string) (bool, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.VaccinationCenters[id]; exists {
		return false, errors.New(`Vaccination Center with ID already exists`)
	}

	vs.VaccinationCenters[id] = VaccinationCenter{
		ID:       id,
		State:    state,
		District: district,
	}

	fmt.Println("VC Data: ", vs.VaccinationCenters)

	return true, nil
}

func AddCapacity(vs *System, centerID string, day string, capacity string) (bool, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.VaccinationCenters[centerID]; !exists {
		return false, errors.New(`Vaccination Center not found`)
	}

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		return false, errors.New(`Invalid day value`)
	}

	capacityInt, err := strconv.Atoi(day)
	if err != nil {
		return false, errors.New(`Invalid capacity value`)
	}

	if vs.Capacity[centerID] == nil {
		vs.Capacity[centerID] = make(map[int]int)
	}

	vs.Capacity[centerID][dayInt] += capacityInt
	return true, nil
}

func BookAppointment(vs *System, centerID string, day string, userID string) (bool, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.VaccinationCenters[centerID]; !exists {
		return false, errors.New(`Vaccination Center not found`)
	}

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		return false, errors.New(`Invalid day value`)
	}

	user, exists := vs.Users[userID]

	if !exists {
		return false, errors.New(`User not found`)
	} else if user.Age <= 18 {
		return false, errors.New(`User is not eligible`)
	}

	maxCapacity := vs.Capacity[centerID][dayInt]

	if maxCapacity <= 0 {
		return false, errors.New(`Cannot Book appointment on given day.`)
	}

	for _, booking := range vs.Appointments[centerID] {
		if booking.UserID == userID && booking.Day == dayInt {
			return false, errors.New(`User already booked an appointment for this day`)
		}
	}

	appointment := Appointment{
		CenterID: centerID,
		UserID:   userID,
		Day:      dayInt,
	}

	vs.Appointments[centerID] = append(vs.Appointments[centerID], appointment)

	vs.Capacity[centerID][dayInt]--

	return true, nil

}

func CancelAppointment(vs *System, centerID string, day string, userID string) (bool, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	_, exists := vs.VaccinationCenters[centerID]
	appointmentIndex := -1
	if !exists {
		return false, errors.New(`Vaccination Center not found`)
	}

	if _, exists := vs.Users[userID]; !exists {
		return false, errors.New(`User not found`)
	}

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		return false, errors.New(`Invalid day value`)
	}

	appointments := vs.Appointments[centerID]

	for i, appointment := range appointments {
		if appointment.UserID == userID && appointment.Day == dayInt {
			appointmentIndex = i
		}
	}

	if appointmentIndex == -1 {
		return false, errors.New(`Appointment not found`)
	}

	vs.Appointments[centerID] = append(vs.Appointments[centerID][:appointmentIndex], vs.Appointments[centerID][appointmentIndex+1:]...)
	vs.Capacity[centerID][dayInt]++

	return true, nil

}

func ListVaccinationCenters(vs *System, district string) []VaccinationCenter {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	result := make([]VaccinationCenter, 0)
	for _, center := range vs.VaccinationCenters {
		if center.District == district {
			result = append(result, center)
		}
	}
	return result
}

func ListAllBookingsOnDay(vs *System, day string, centerId string) ([]Appointment, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	_, exists := vs.VaccinationCenters[centerId]
	if !exists {
		return nil, errors.New(`Vaccination Center with ID does not exist`)
	}

	dayInt, err := strconv.Atoi(day)
	if err != nil {
		return nil, errors.New(`Invalid day value`)
	}

	appointments := vs.Appointments[centerId]

	if len(appointments) == 0 {
		return nil, errors.New(`'No Bookings found for given center`)
	}

	result := make([]Appointment, 0)

	for _, appointment := range appointments {
		if appointment.Day == dayInt {
			result = append(result, appointment)
		}
	}

	if len(result) == 0 {
		return nil, errors.New(`No bookings found for day`)
	}

	return result, nil
}

func handleCommands(command []string, vs *System) {
	if len(command) == 0 || command[0] == "" {
		fmt.Println("Invalid instruction: Command is empty")
		return
	}
	switch command[0] {
	case "ADD_USER":
		_, err := AddUser(vs, command[1], command[2], command[3], command[4], command[5], command[6])

		if err != nil {
			fmt.Println(`Error in AddUser: `, err)
		}

		fmt.Println(`Action Success: `, vs.Users[command[1]])
	case "ADD_VACCINATION_CENTER":
		_, err := AddVC(vs, command[1], command[2], command[3])

		if err != nil {
			fmt.Println(`Error in Add Vaccination Center: `, err)
		}

		fmt.Println(`Action Success: `, vs.VaccinationCenters[command[4]])

	case "ADD_CAPACITY":
		_, err := AddCapacity(vs, command[1], command[2], command[3])

		if err != nil {
			fmt.Println(`Error in Add Capacity: `, err)
		}

		fmt.Println(`Action Success: `, vs.Capacity[command[1]])

	case "LIST_VACCINATION_CENTERS":
		data := ListVaccinationCenters(vs, command[1])

		if len(data) == 0 {
			fmt.Println(`No Vaccination Centers found for district`)
		}

		fmt.Println(`Action Success: `, data)

	case "CANCEL_BOOKING":
		_, err := CancelAppointment(vs, command[1], command[2], command[3])

		if err != nil {
			fmt.Println(`Error while Cancelling Booking: `, err)
		}

		fmt.Println(`Action Success: Booking cancelled`, vs.Appointments[command[1]])

	case "BOOK_VACCINATION":
		_, err := BookAppointment(vs, command[1], command[2], command[3])

		if err != nil {
			fmt.Println(`Error occured while booking: `, err)
		}

		fmt.Println(`Action Success: Booking Confirmed `, vs.Appointments[command[1]])

	case "LIST_ALL_BOOKINGS":
		data, err := ListAllBookingsOnDay(vs, command[1], command[2])

		if err != nil {
			fmt.Println(`Error in fetching bookings for mentioned day: `, err)
		}

		fmt.Println(`Action Success: `, data)

	default:
		fmt.Printf("Invalid command")
	}
}

func main() {

	vs := System{
		Users:              make(map[string]User),
		VaccinationCenters: make(map[string]VaccinationCenter),
		Capacity:           make(map[string]map[int]int),
		Appointments:       make(map[string][]Appointment),
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter instruction \t")
		scanner.Scan()
		command := scanner.Text()
		actionDetails := strings.Fields(command)
		handleCommands(actionDetails, &vs)
	}
}
