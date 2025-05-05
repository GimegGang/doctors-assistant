package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"kode/internal/component/reception"
	"kode/internal/storage"
	"net/http"
	"slices"
	"testing"
	"time"
)

func TestRest(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       "../",
			Dockerfile:    "main.dockerfile",
			KeepImage:     false,
			PrintBuildLog: true,
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	if logs, err := container.Logs(ctx); err == nil {
		t.Logf("Container logs:\n%s", logs)
	}

	port, err := container.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	endpoint := "http://localhost:" + port.Port()
	time.Sleep(2 * time.Second)

	requestPostData := storage.Medicine{
		Name:              "test",
		TakingDuration:    5,
		TreatmentDuration: 5,
		UserId:            1,
	}

	var createdID int64

	t.Run("POST /schedule", func(t *testing.T) {
		jsonBody, _ := json.Marshal(requestPostData)
		resp, err := http.Post(
			endpoint+"/schedule",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			t.Fatalf("POST failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body := new(bytes.Buffer)
			body.ReadFrom(resp.Body)
			t.Fatalf("Expected 200, got %d. Response body: %s", resp.StatusCode, body.String())
		}

		var response struct {
			Id int64 `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if response.Id <= 0 {
			t.Fatalf("Expected positive ID, got %d", response.Id)
		}

		createdID = response.Id
	})

	t.Run("GET /schedule", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No schedule ID available (POST test might have failed)")
		}

		req, err := http.NewRequest("GET", endpoint+"/schedule?user_id=1&schedule_id="+fmt.Sprint(createdID), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body := new(bytes.Buffer)
			body.ReadFrom(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, body.String())
		}

		var result storage.Medicine
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Проверка полей
		if result.Name != requestPostData.Name {
			t.Errorf("Expected Name %q, got %q", requestPostData.Name, result.Name)
		}
		if result.TreatmentDuration != requestPostData.TreatmentDuration {
			t.Errorf("Expected TreatmentDuration %d, got %d", requestPostData.TreatmentDuration, result.TreatmentDuration)
		}
		if result.TakingDuration != requestPostData.TakingDuration {
			t.Errorf("Expected TakingDuration %d, got %d", requestPostData.TakingDuration, result.TakingDuration)
		}
		if result.UserId != requestPostData.UserId {
			t.Errorf("Expected UserId %d, got %d", requestPostData.UserId, result.UserId)
		}

		expectedSchedule, err := reception.GetReceptionIntake(result.TakingDuration)
		if err != nil {
			t.Fatalf("Failed to generate expected schedule: %v", err)
		}
		if !slices.Equal(result.Schedule, expectedSchedule) {
			t.Errorf("Expected Schedule %v, got %v", expectedSchedule, result.Schedule)
		}
	})

	t.Run("GET /schedules", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No schedule ID available (POST test might have failed)")
		}

		req, err := http.NewRequest(
			"GET",
			endpoint+"/schedules?user_id="+fmt.Sprint(requestPostData.UserId),
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body := new(bytes.Buffer)
			body.ReadFrom(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, body.String())
		}

		var result struct {
			SchedulesID []int64 `json:"schedules_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(result.SchedulesID) != 1 {
			t.Errorf("Expected len %d, got %d", 1, len(result.SchedulesID))
		}
		if result.SchedulesID[0] != createdID {
			t.Errorf("Expected schedule id %d, got %d", createdID, result.SchedulesID[0])
		}
	})

	t.Run("GET /next_takings", func(t *testing.T) {
		req, err := http.NewRequest(
			"GET",
			endpoint+"/next_takings?user_id="+fmt.Sprint(requestPostData.UserId),
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body := new(bytes.Buffer)
			body.ReadFrom(resp.Body)
			t.Fatalf("Expected status 200, got %d. Response: %s", resp.StatusCode, body.String())
		}

		//TODO дописать проверку хендлера
	})
}
