package generator

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const binary = "./generator/dntgenerator"

func GenerateLevel(start int32, end int32, max int32) string {
	cmd := exec.Command(binary, "-s", strconv.Itoa(int(start)), "-e", strconv.Itoa(int(end)), "-m", strconv.Itoa(int(max)), "-j", "-", "-h", "")

	stderr := &strings.Builder{}
	stdout := &strings.Builder{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	log.Info().Msgf("raw generator output: %s", stdout.String())

	if err := cmd.Run(); err != nil {
		log.Warn().Msgf("stderr: %s", stderr.String())
		log.Fatal().Msgf("failed to run %s: %v", binary, err)
	}

	return stdout.String()
}
