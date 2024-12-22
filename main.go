package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"html/template"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// Initialize MySQL connection
func init() {
	dbUser := "root"
	dbPassword := "linux" // Change to your actual MySQL password
	dbName := "student_attendance"
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPassword, dbName)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}

// Structure to hold attendance record
type Attendance struct {
	ID               int
	FirstName        string
	LastName         string
	DOB              string
	Phone            string
	Address          string
	CourseName       string
	CourseLength     string
	GraduationDate   string
	TutorName        string
	TimeIn           string
	TimeOut          string
	AbsenceReason    string
	CameraStatus     string
}

// Render the form and list
func renderForm(w http.ResponseWriter, r *http.Request) {
	attendanceList, err := getAllAttendance()
	if err != nil {
		http.Error(w, "Unable to fetch attendance", http.StatusInternalServerError)
		return
	}

	// Checking if there is an ID in the URL for editing purposes
	id := r.URL.Query().Get("id")
	var editAttendance Attendance
	if id != "" {
		editID, err := strconv.Atoi(id)
		if err == nil {
			editAttendance, err = getAttendanceByID(editID)
			if err != nil {
				http.Error(w, "Unable to fetch the attendance record", http.StatusInternalServerError)
				return
			}
		}
	}

	data := struct {
		Attendance  []Attendance // All records for the list
		EditRecord Attendance    // The record to be edited
	}{
		Attendance: attendanceList,
		EditRecord: editAttendance,
	}
	renderTemplate(w, "index", data)
}

// Function to render templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := fmt.Sprintf("templates/%s.html", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

// Fetch all attendance records from the database
func getAllAttendance() ([]Attendance, error) {
	rows, err := db.Query("SELECT id, first_name, last_name, dob, phone, address, course_name, course_length, graduation_date, tutor_name, time_in, time_out, absence_reason, camera_status FROM attendance")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendanceList []Attendance
	for rows.Next() {
		var attendance Attendance
		err := rows.Scan(&attendance.ID, &attendance.FirstName, &attendance.LastName, &attendance.DOB, &attendance.Phone, &attendance.Address, &attendance.CourseName, &attendance.CourseLength, &attendance.GraduationDate, &attendance.TutorName, &attendance.TimeIn, &attendance.TimeOut, &attendance.AbsenceReason, &attendance.CameraStatus)
		if err != nil {
			return nil, err
		}
		attendanceList = append(attendanceList, attendance)
	}

	return attendanceList, nil
}

// Fetch a single attendance record by ID
func getAttendanceByID(id int) (Attendance, error) {
	var attendance Attendance
	err := db.QueryRow("SELECT id, first_name, last_name, dob, phone, address, course_name, course_length, graduation_date, tutor_name, time_in, time_out, absence_reason, camera_status FROM attendance WHERE id=?", id).Scan(
		&attendance.ID, &attendance.FirstName, &attendance.LastName, &attendance.DOB, &attendance.Phone, &attendance.Address, 
		&attendance.CourseName, &attendance.CourseLength, &attendance.GraduationDate, &attendance.TutorName, &attendance.TimeIn, 
		&attendance.TimeOut, &attendance.AbsenceReason, &attendance.CameraStatus,
	)
	if err != nil {
		return attendance, err
	}
	return attendance, nil
}

// Add or update attendance in the database
func addOrUpdateAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form data", http.StatusBadRequest)
			return
		}

		// Collect form values
		id, _ := strconv.Atoi(r.FormValue("id"))
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		dob := r.FormValue("dob")
		phone := r.FormValue("phone")
		address := r.FormValue("address")
		courseName := r.FormValue("courseName")
		courseLength := r.FormValue("courseLength")
		graduationDate := r.FormValue("graduationDate")
		tutorName := r.FormValue("tutorName")
		timeIn := r.FormValue("timeIn")
		timeOut := r.FormValue("timeOut")
		absenceReason := r.FormValue("absenceReason")
		cameraStatus := r.FormValue("cameraStatus")

		// Add or update attendance
		if id == 0 {
			_, err = db.Exec(`
				INSERT INTO attendance (first_name, last_name, dob, phone, address, course_name, course_length, graduation_date, tutor_name, time_in, time_out, absence_reason, camera_status)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				firstName, lastName, dob, phone, address, courseName, courseLength, graduationDate, tutorName, timeIn, timeOut, absenceReason, cameraStatus)
		} else {
			_, err = db.Exec(`
				UPDATE attendance SET first_name=?, last_name=?, dob=?, phone=?, address=?, course_name=?, course_length=?, graduation_date=?, tutor_name=?, time_in=?, time_out=?, absence_reason=?, camera_status=?
				WHERE id=?`,
				firstName, lastName, dob, phone, address, courseName, courseLength, graduationDate, tutorName, timeIn, timeOut, absenceReason, cameraStatus, id)
		}

		if err != nil {
			http.Error(w, "Failed to save attendance", http.StatusInternalServerError)
			return
		}

		// Redirect back to the form with updated data
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Delete an attendance record
func deleteAttendance(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM attendance WHERE id=?", id)
	if err != nil {
		http.Error(w, "Failed to delete attendance", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handle static file requests (CSS, JS)
func handleStaticFiles() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func main() {
	// Serve static files (CSS)
	handleStaticFiles()

	// Routes
	http.HandleFunc("/", renderForm)                // Route to render form (index.html)
	http.HandleFunc("/add", addOrUpdateAttendance)  // Route to add/update attendance
	http.HandleFunc("/delete", deleteAttendance)    // Route to delete attendance

	port := ":8080"
	fmt.Println("Server started at http://localhost" + port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
