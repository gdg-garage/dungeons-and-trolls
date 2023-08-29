package generator

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const binary = "./generator/dntgenerator"

func Generate_level(start int, end int, max int) string {
	cmd := exec.Command(binary, "-s", strconv.Itoa(start), "-e", strconv.Itoa(end), "-m", strconv.Itoa(max), "-j", "-", "-h", "")

	stderr := &strings.Builder{}
	stdout := &strings.Builder{}
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		log.Warn().Msgf("stderr: %s", stderr.String())
		log.Fatal().Msgf("failed to run %s: %v", binary, err)
	}

	return stdout.String()
}
