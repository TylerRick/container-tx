package compose

import (
	"os"
	"testing"
)

func TestIngestContainers(t *testing.T) {
	cf := DockerCompose{}

	f, err := os.Open("./test_fixtures/docker-compose.yaml")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	_, err = cf.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}
}

func TestEmitContainers(t *testing.T) {
	cf := DockerCompose{}

	f, err := os.Open("./test_fixtures/docker-compose.yaml")
	if err != nil {
		t.Errorf("Failed to open fixture: %s", err)
	}

	bp, err := cf.IngestContainers(f)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

	_, err = DockerCompose{}.EmitContainers(bp)
	if err != nil {
		t.Errorf("Failed to ingest containers: %s", err)
	}

}
