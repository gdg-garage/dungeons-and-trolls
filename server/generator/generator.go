package generator

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/rs/zerolog/log"
)

const binary = "./generator/dntgenerator"
const outfile = "/tmp/level.json"

func Generate_level(start int, end int, max int) {
	cmd := exec.Command(binary, "-s", strconv.Itoa(start), "-e", strconv.Itoa(end), "-m", strconv.Itoa(max), "-j", outfile)
	// res, err := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Println(string(res))
	// 	panic(err)
	// }
	// fmt.Println(string(res))

	// TODO show stderr
	if err := cmd.Run(); err != nil {
		log.Fatal().Msgf("failed to run %s: %v", binary, err)
	}
	dat, err := os.ReadFile(outfile)
	if err != nil {
		log.Fatal().Msgf("failed to read generated file %s: %v", outfile, err)
	}
	fmt.Print(string(dat))
}
