package main

import (
	g "github.com/AllenDang/giu"
	"github.com/BurntSushi/toml"

	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var (
	// project settings
	projectPWD string

	// plugin manifest
	pluginID      string
	pluginName    string
	pluginAuthor  string
	pluginVersion string

	// build for settings
	linuxAmd64   bool = true
	linuxArm64   bool = true
	windowsAmd64 bool = true
	darwinAmd64  bool = true

	// marketplace
	marketplaceRepoDir string
)

// pluginDescription is a plugin manifest file.
type pluginDescription struct {
	ID      string `toml:"ID"`
	Name    string `toml:"NAME"`
	Author  string `toml:"AUTHOR"`
	Version string `toml:"VERSION"`
}

func loadManifest() {
	var err error
	var manifest pluginDescription
	if _, err = toml.DecodeFile(projectPWD+"/plugin.toml", &manifest); err != nil {
		fmt.Println("Cannot load plugin.toml manifest!")
		return
	}

	pluginID = manifest.ID
	pluginName = manifest.Name
	pluginAuthor = manifest.Author
	pluginVersion = manifest.Version
}

func buildPackage() {
	tempDir, err := ioutil.TempDir("", "iglu-package")
	if err != nil {
		fmt.Println(err)
		return
	}

	if linuxAmd64 {
		buildBinary(tempDir, "linux", "amd64")
	}
	if linuxArm64 {
		buildBinary(tempDir, "linux", "arm64")
	}
	if windowsAmd64 {
		buildBinary(tempDir, "windows", "amd64")
	}
	if darwinAmd64 {
		buildBinary(tempDir, "darwin", "amd64")
	}

}

func buildBinary(tempDir, goos, goarch string) {
	fmt.Printf("Building for %s/%s\n", goos, goarch)
	path := fmt.Sprintf("%s/%s-%s", tempDir, goos, goarch)
	binaryPath := fmt.Sprintf("%s/%s", path, pluginID)
	binaryPathXZ := fmt.Sprintf("%s/%s.xz", path, pluginID)
	manifestPath := fmt.Sprintf("%s/%s.toml", path, pluginID)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath)
	cmd.Dir = projectPWD
	env := os.Environ()
	env = append(env, "GOOS="+goos)
	env = append(env, "GOARCH="+goarch)
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Go build failed: ", out)
		return
	}

	// Archive the binary (with xz)
	cmdXz := exec.Command("xz", binaryPath)
	out, err = cmdXz.CombinedOutput()
	if err != nil {
		fmt.Println("xz failed: ", out)
		return
	}

	// Copy the manifest
	manifestData, err := ioutil.ReadFile("./plugin.toml")
	if err != nil {
		fmt.Println("loading manifest failed: ", err)
		return
	}
	err = ioutil.WriteFile(manifestPath, manifestData, 0644)
	if err != nil {
		fmt.Println("saving manifest failed: ", err)
		return
	}

	// Calculate sha256
	cmdSha := exec.Command("sha256sum", binaryPathXZ)
	out, err = cmdSha.CombinedOutput()
	if err != nil {
		fmt.Println("sha failed: ", out)
		return
	}

}

func loop() {
	g.SingleWindow("packager", g.Layout{
		g.Label("iglü plugin packager"),
		g.InputText("Plugin", 600, &projectPWD),
		g.Button("Load Manifest", loadManifest),
		g.Label("Plugin Manifest (read-only)"),
		g.InputTextV("ID", 250, &pluginID, g.InputTextFlagsReadOnly, nil, func() {}),
		g.InputTextV("Name", 250, &pluginName, g.InputTextFlagsReadOnly, nil, func() {}),
		g.InputTextV("Author", 250, &pluginAuthor, g.InputTextFlagsReadOnly, nil, func() {}),
		g.InputTextV("Version", 250, &pluginVersion, g.InputTextFlagsReadOnly, nil, func() {}),
		g.Label("Build for..."),
		g.Line(
			g.Checkbox("linux/amd64", &linuxAmd64, func() {}),
			g.Checkbox("linux/arm64", &linuxArm64, func() {}),
			g.Checkbox("windows/amd64", &windowsAmd64, func() {}),
			g.Checkbox("darwin/amd64", &darwinAmd64, func() {}),
		),
		g.InputText("Repo Dir", 600, &marketplaceRepoDir),
		g.Line(
			g.ButtonV("Package", -1, -1, buildPackage)),
	})
}

func main() {
	var err error
	projectPWD, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	loadManifest()

	marketplaceRepoDir = os.Getenv("IGLU_MARKETPLACE")
	wnd := g.NewMasterWindow("iglü plugin packager", 700, 300, g.MasterWindowFlagsNotResizable, nil)
	wnd.Main(loop)
}
