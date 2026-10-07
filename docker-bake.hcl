group "default" {
  targets = [
    "linux-amd64",
    "windows-amd64"
  ]
}

target "setup-env" {
  context    = "."

  dockerfile = ".bake/setup-env.dockerfile"

  platforms  = ["linux/amd64"]
}

target "setup-workspace" {
  context    = "."

  dockerfile = ".bake/setup-workspace.dockerfile"

  platforms  = ["linux/amd64"]

  contexts = {
    build-env = "target:setup-env"
  }
}

target "build" {
  context    = "."

  dockerfile = ".bake/build.dockerfile"

  target     = "artifact"

  platforms  = ["linux/amd64"]

  contexts = {
    workspace = "target:setup-workspace"
  }
}

target "linux-amd64" {
  inherits = ["build"]

  args = {
    GOOS        = "linux"
    GOARCH      = "amd64"
    BINARY_NAME = "fit"
  }

  output = ["type=local,dest=dist/linux-amd64,platform-split=false"]
}

target "windows-amd64" {
  inherits = ["build"]

  args = {
    GOOS        = "windows"
    GOARCH      = "amd64"
    BINARY_NAME = "fit.exe"
  }

  output = ["type=local,dest=dist/windows-amd64,platform-split=false"]
}
