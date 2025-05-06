package tests

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kode/internal/component/reception"
	medicineProto "kode/proto/gen"
	"slices"
	"testing"
	"time"
)

func TestGRPC(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       "../",
			Dockerfile:    "main.dockerfile",
			KeepImage:     false,
			PrintBuildLog: true,
		},
		ExposedPorts: []string{"1234/tcp"},
		WaitingFor:   wait.ForListeningPort("1234/tcp").WithStartupTimeout(30 * time.Second),
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

	port, err := container.MappedPort(ctx, "1234")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	conn, err := grpc.NewClient(
		"localhost:"+port.Port(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	client := medicineProto.NewMedicineServiceClient(conn)

	requestPostData := medicineProto.AddScheduleRequest{
		Name:              "test",
		TakingDuration:    5,
		TreatmentDuration: 5,
		UserId:            1,
	}

	var createdID int64

	t.Run("gRPC AddSchedule", func(t *testing.T) {
		response, err := client.AddSchedule(ctx, &requestPostData)
		if err != nil {
			t.Fatalf("Error request: %v", err)
		}
		if response.Id <= 0 {
			t.Fatalf("Expected positive ID, got %d", response.Id)
		}
		createdID = response.Id
	})

	t.Run("gRPC Schedule", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No schedule ID available")
		}

		response, err := client.Schedule(ctx, &medicineProto.ScheduleRequest{
			ScheduleId: createdID,
			UserId:     requestPostData.UserId,
		})
		if err != nil {
			t.Fatalf("Error request: %v", err)
		}

		if response.Name != requestPostData.Name {
			t.Errorf("Expected Name %q, got %q", requestPostData.Name, response.Name)
		}
		if response.TreatmentDuration != requestPostData.TreatmentDuration {
			t.Errorf("Expected TreatmentDuration %d, got %d", requestPostData.TreatmentDuration, response.TreatmentDuration)
		}
		if response.TakingDuration != requestPostData.TakingDuration {
			t.Errorf("Expected TakingDuration %d, got %d", requestPostData.TakingDuration, response.TakingDuration)
		}
		if response.UserId != requestPostData.UserId {
			t.Errorf("Expected UserId %d, got %d", requestPostData.UserId, response.UserId)
		}

		expectedSchedule, err := reception.GetReceptionIntake(response.TakingDuration)
		if err != nil {
			t.Fatalf("Failed to generate expected schedule: %v", err)
		}
		if !slices.Equal(response.Schedule, expectedSchedule) {
			t.Errorf("Expected Schedule %v, got %v", expectedSchedule, response.Schedule)
		}
	})

	t.Run("gRPC Schedules", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("No schedule ID available")
		}

		response, err := client.Schedules(ctx, &medicineProto.SchedulesRequest{UserId: requestPostData.UserId})
		if err != nil {
			t.Fatalf("Error request: %v", err)
		}

		if len(response.SchedulesId) != 1 {
			t.Errorf("Expected len %d, got %d", 1, len(response.SchedulesId))
		}
		if response.SchedulesId[0] != createdID {
			t.Errorf("Expected schedule id %d, got %d", createdID, response.SchedulesId[0])
		}
	})

	t.Run("gRPC NextTakings", func(t *testing.T) {
		_, err := client.NextTakings(ctx, &medicineProto.NextTakingsRequest{UserId: requestPostData.UserId})
		if err != nil {
			t.Fatalf("Error request: %v", err)
		}

	})
}
