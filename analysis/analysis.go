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

func AnalyzePythonRequirementsFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}

	defer file.Close()
	reader := bufio.NewReader(file)

	packageNames := dependencies.ParsePythonRequirements(reader)
	result := false

	for _, pythonPackageName := range packageNames {
		availableForRegistration, err := registry.IsPypiPackageAvailableForRegistration(pythonPackageName)

		if err != nil {
			fmt.Println(err)
			return false, err
		}

		if availableForRegistration {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("[!] python package \"%s\" is available for public registration. %s", pythonPackageName, filePath))
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

	for _, packageName := range packageNames {
		availableForRegistration, err := registry.IsNpmPackageAvailableForRegistration(packageName)

		if err != nil {
			fmt.Println(err)
			return false, err
		}

		if availableForRegistration {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("[!] npm package \"%s\" is available for public registration. %s", packageName, filePath))
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
		// Define regex patterns for filenames
		filePatterns := []string{
			`package.*\.json`,
			`yarn.*\.json`,
			`requirements.*\.txt`,
		}
		fileName := fileInfo.Name()
		for pattern_index, pattern := range filePatterns {
			if matched, _ := regexp.MatchString(pattern, fileName); matched {
				if pattern_index == 0 || pattern_index == 1 {
					fmt.Printf("scanning \"%v\"\n", path)
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

		if fileName == "package.json" {
			result, err := AnalyzePackagesJsonFile(filePath)
			if result {
				hasAnyPackageAvailableForRegistration = true
			}

			if err != nil {
				return false, err
			}
		}

		if fileName == "requirements.txt" {
			result, err := AnalyzePythonRequirementsFile(filePath)
			if result {
				hasAnyPackageAvailableForRegistration = true
			}

			if err != nil {
				return false, err
			}
		}
	}

	return hasAnyPackageAvailableForRegistration, nil
}
