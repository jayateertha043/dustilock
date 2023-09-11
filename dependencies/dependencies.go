package dependencies

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func ParsePythonRequirements(reader *bufio.Reader) map[string]string {

	packageVersions := make(map[string]string)
	for {
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		line := string(lineBytes)
		line = strings.TrimSpace(line)

		re := regexp.MustCompile(`[#&]+egg=([a-zA-Z0-9_\-.]+)`)
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			packageName := strings.ToLower(match[1])
			packageVersions[packageName] = "any"
			continue
		}

		line = strings.Split(line, "#")[0]

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "--requirement") {
			continue
		}

		if strings.HasPrefix(line, "-r") {
			continue
		}

		if strings.Contains(line, "://") {
			continue
		}

		re = regexp.MustCompile(`^([a-zA-Z0-9_\-.]+)(?:([=,>,<])?=?([0-9.]+))?`)
		match = re.FindStringSubmatch(line)
		if len(match) > 0 {
			packageName := strings.ToLower(match[1])
			operator := ""
			if len(match) > 3 {
				operator = match[2]
				packageVersion := operator + match[3]
				if operator == "=" {
					packageVersion = packageVersion + ":conda"
				}
				packageVersions[packageName] = packageVersion
			} else {
				packageVersions[packageName] = "*"
			}
			continue
		}

	}
	return packageVersions
}

func ParsePackagesJsonFile(reader *bufio.Reader) (map[string]string, error) {
	packageVersions := make(map[string]string)

	d := json.NewDecoder(reader)
	t := struct {
		Dependencies    *map[string]interface{} `json:"dependencies"`
		DevDependencies *map[string]interface{} `json:"devDependencies"`
	}{}

	err := d.Decode(&t)
	if err != nil {
		return nil, err
	}

	processPackageName := func(dict *map[string]interface{}, npmPackageName string) {
		if strings.HasPrefix(npmPackageName, "@") {
			return
		}

		value, ok := (*dict)[npmPackageName]
		if !ok {
			packageVersions[npmPackageName] = "*" // Assign "*" if version is not specified
			return
		}

		version := fmt.Sprintf("%v", value)
		version = strings.ToLower(version)

		if strings.HasPrefix(version, "npm:") {
			return
		}

		if strings.Contains(version, "://") {
			return
		}

		packageVersions[npmPackageName] = version
	}

	if t.Dependencies != nil {
		for npmPackageName := range *t.Dependencies {
			processPackageName(t.Dependencies, npmPackageName)
		}
	}

	if t.DevDependencies != nil {
		for npmPackageName := range *t.DevDependencies {
			processPackageName(t.DevDependencies, npmPackageName)
		}
	}

	return packageVersions, nil
}
