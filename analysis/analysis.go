package analysis

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/jayateertha043/dustilock/dependencies"
	"github.com/jayateertha043/dustilock/registry"
)

// Define regex patterns for filenames
var filePatterns = []string{
	`package.*\.json`,
	`yarn.*\.json`,
	`requirements.*\.txt`,
}

func AnalyzePythonRequirementsFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}

	defer file.Close()
	reader := bufio.NewReader(file)

	packageNames := dependencies.ParsePythonRequirements(reader)
	result := false

	for pythonPackageName, pythonPackageVersion := range packageNames {
		if pythonPackageName == "" {
			continue
		}
		availableForRegistration, err := registry.IsPypiPackageAvailableForRegistration(pythonPackageName)

		if err != nil {
			fmt.Println(err)
			return false, err
		}

		if availableForRegistration {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("[!] python package \"%s:%s\" is available for public registration. %s", pythonPackageName, pythonPackageVersion, filePath))
			result = true
		}
	}

	return result, nil
}

func AnalyzePackagesJsonFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}

	defer file.Close()
	reader := bufio.NewReader(file)
	packageNames, err := dependencies.ParsePackagesJsonFile(reader)
	if err != nil {
		return false, err
	}

	result := false

	for npmpackageName, npmPackageVersion := range packageNames {
		availableForRegistration, err := registry.IsNpmPackageAvailableForRegistration(npmpackageName)

		if err != nil {
			fmt.Println(err)
			return false, err
		}

		if availableForRegistration {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("[!] npm package \"%s:%s\" is available for public registration. %s", npmpackageName, npmPackageVersion, filePath))
			result = true
		}
	}

	return result, nil
}

func AnalyzeDirectoryRecursive(workingDir string, excludedDirectories map[string]bool) (bool, error) {
	hasAnyPackageAvailableForRegistration := false

	err := filepath.Walk(workingDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if excludedDirectories != nil {
			_, isExcluded := excludedDirectories[fileInfo.Name()]
			if fileInfo.IsDir() && isExcluded {
				return filepath.SkipDir
			}
		}

		fileName := fileInfo.Name()
		for pattern_index, pattern := range filePatterns {
			if matched, _ := regexp.MatchString(pattern, fileName); matched {
				fmt.Printf("[*] Scanning \"%v\"\n", path)
				if pattern_index == 0 || pattern_index == 1 {
					result, err := AnalyzePackagesJsonFile(path)
					if result {
						hasAnyPackageAvailableForRegistration = true
					}

					if err != nil {
						fmt.Println(err)
					}
					return err
				} else if pattern_index == 2 {
					result, err := AnalyzePythonRequirementsFile(path)
					if result {
						hasAnyPackageAvailableForRegistration = true
					}

					if err != nil {
						fmt.Println(err)
					}
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	return hasAnyPackageAvailableForRegistration, nil
}

func AnalyzeDirectory(workingDir string) (bool, error) {
	hasAnyPackageAvailableForRegistration := false

	files, err := ioutil.ReadDir(workingDir)
	if err != nil {
		return false, err
	}

	for _, fileInfo := range files {
		fileName := fileInfo.Name()
		filePath := path.Join(workingDir, fileName)
		for pattern_index, pattern := range filePatterns {
			if matched, _ := regexp.MatchString(pattern, fileName); matched {
				fmt.Printf("[*] Scanning \"%v\"\n", filePath)
				if pattern_index == 0 || pattern_index == 1 {
					result, err := AnalyzePackagesJsonFile(filePath)
					if result {
						hasAnyPackageAvailableForRegistration = true
					}
					if err != nil {
						fmt.Println(err)
						return false, err
					}
				} else if pattern_index == 2 {
					result, err := AnalyzePythonRequirementsFile(filePath)
					if result {
						hasAnyPackageAvailableForRegistration = true
					}
					if err != nil {
						fmt.Println(err)
						return false, err
					}
				}
			}
		}
	}

	return hasAnyPackageAvailableForRegistration, nil
}
