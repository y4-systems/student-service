package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// EnrollmentClient handles communication with the Enrollment Service
type EnrollmentClient struct {
	baseURL string
	client  *http.Client
}

// Enrollment represents an enrollment record from the Enrollment Service
type Enrollment struct {
	ID        string    `json:"_id"`
	StudentID string    `json:"student_id"`
	CourseID  string    `json:"course_id"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// StudentEnrollments represents all enrollments for a student
type StudentEnrollments struct {
	StudentID   string         `json:"student_id"`
	Name        string         `json:"name,omitempty"`
	Email       string         `json:"email,omitempty"`
	Enrollments []Enrollment   `json:"enrollments"`
	Count       int            `json:"count"`
}

// NewEnrollmentClient creates a new Enrollment Service client
func NewEnrollmentClient() *EnrollmentClient {
	baseURL := os.Getenv("ENROLLMENT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5003"
	}

	return &EnrollmentClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetStudentEnrollments retrieves all enrollments for a specific student from the Enrollment Service
func (ec *EnrollmentClient) GetStudentEnrollments(studentID string) (*StudentEnrollments, error) {
	url := fmt.Sprintf("%s/enrollments/student/%s", ec.baseURL, studentID)

	resp, err := ec.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call enrollment service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read enrollment service response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			// Student has no enrollments - return empty list instead of error
			return &StudentEnrollments{
				StudentID:   studentID,
				Enrollments: []Enrollment{},
				Count:       0,
			}, nil
		}
		return nil, fmt.Errorf("enrollment service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - could be array or wrapped object
	var enrollments []Enrollment
	if err := json.Unmarshal(body, &enrollments); err != nil {
		// Try parsing as wrapped response
		var wrapped struct {
			Enrollments []Enrollment `json:"enrollments"`
		}
		if err := json.Unmarshal(body, &wrapped); err != nil {
			return nil, fmt.Errorf("failed to parse enrollment service response: %w", err)
		}
		enrollments = wrapped.Enrollments
	}

	return &StudentEnrollments{
		StudentID:   studentID,
		Enrollments: enrollments,
		Count:       len(enrollments),
	}, nil
}

// GetCourseRoster retrieves all students enrolled in a specific course
func (ec *EnrollmentClient) GetCourseRoster(courseID string) ([]Enrollment, error) {
	url := fmt.Sprintf("%s/enrollments/course/%s", ec.baseURL, courseID)

	resp, err := ec.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call enrollment service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read enrollment service response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return []Enrollment{}, nil
		}
		return nil, fmt.Errorf("enrollment service returned status %d: %s", resp.StatusCode, string(body))
	}

	var enrollments []Enrollment
	if err := json.Unmarshal(body, &enrollments); err != nil {
		// Try parsing as wrapped response
		var wrapped struct {
			Enrollments []Enrollment `json:"enrollments"`
		}
		if err := json.Unmarshal(body, &wrapped); err != nil {
			return nil, fmt.Errorf("failed to parse enrollment service response: %w", err)
		}
		enrollments = wrapped.Enrollments
	}

	return enrollments, nil
}

// CheckEnrollmentStatus checks if a student is enrolled in a course
func (ec *EnrollmentClient) CheckEnrollmentStatus(studentID, courseID string) (bool, error) {
	url := fmt.Sprintf("%s/enrollments/check?student_id=%s&course_id=%s", ec.baseURL, studentID, courseID)

	resp, err := ec.client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to call enrollment service: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Health checks if the Enrollment Service is reachable
func (ec *EnrollmentClient) Health() (bool, error) {
	url := fmt.Sprintf("%s/", ec.baseURL)

	resp, err := ec.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
