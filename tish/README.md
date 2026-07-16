# Tiny Shell (`tish`)

`tish` (Tiny Shell) is a lightweight, interactive shell written in Go from scratch. It is designed to demonstrate key systems programming concepts, including process lifecycle management (`fork`/`exec`), standard I/O redirection, command pipelines, background job management, and POSIX signal handling.

---

## 🚀 Features

- **Interactive REPL**: Command execution loop with custom prompt (`tish>`).
- **External Commands**: Executes standard system binaries available in `$PATH`.
- **Built-in Commands**:
  - `cd <directory>`: Changes current working directory.
  - `pwd`: Prints current working directory.
  - `export KEY=VALUE`: Sets shell environment variables.
  - `jobs`: Lists active background process PIDs.
  - `wait`: Waits for all active background jobs to complete.
  - `exit`: Gracefully exits the shell.
- **Environment Variable Expansion**: Supports `$VAR` and `${VAR}` syntax inside command lines, as well as prefix overrides (`VAR=val cmd`).
- **I/O Redirection**:
  - `< file`: Redirects standard input (`stdin`) from a file.
  - `> file`: Redirects standard output (`stdout`) to a file (truncates existing file).
  - `>> file`: Redirects standard output (`stdout`) to a file (appends to existing file).
- **Pipelines (`|`)**: Connects multiple processes via kernel pipes (`os.Pipe`), streaming output from one command directly into the next.
- **Background Execution (`&`)**: Spawns processes in isolated process groups (`Setpgid`) so they run concurrently without blocking the shell.
- **Signal Handling**: Intercepts `Ctrl+C` (`SIGINT`) safely without killing background processes or crashing the shell.

---

## 🛠️ Getting Started

### Prerequisites

- **OS**: Linux or WSL (Windows Subsystem for Linux).
- **Go**: Version 1.22 or higher recommended.

### Run the Shell

To run the shell directly:

```bash
cd tish
go run main.go
```

### Build the Binary

To compile the shell into an executable binary:

```bash
cd tish
go build -o tish main.go
./tish
```

---

## ⌨️ Arrow Keys & Command History (`rlwrap`)

By default, standard terminal input streams (`bufio.Scanner`) do not process line-editing keybindings like Arrow Keys (Up/Down for history, Left/Right for moving the cursor) and will display raw ANSI escape codes (such as `^[[A` or `^[[B`).

To enable **Command History (Up/Down arrows)**, **Line Editing (Left/Right arrows)**, and **Ctrl+A / Ctrl+E navigation**, wrap `tish` with `rlwrap` (Readline Wrapper):

### Installing `rlwrap`:
- **Ubuntu/Debian/WSL**: `sudo apt install rlwrap`
- **Fedora**: `sudo dnf install rlwrap`
- **Arch**: `sudo pacman -S rlwrap`

### Running with `rlwrap`:
```bash
rlwrap ./tish
```
or during development:
```bash
rlwrap go run main.go
```

---

## ⚠️ OS & Platform Compatibility

> [!WARNING]
> `tish` is specifically built and tested for **UNIX/Linux environments** (including WSL).

- **Linux / WSL**: **Fully Supported.** All POSIX process group controls (`Setpgid`), signal delivery (`SIGINT`), and file descriptor pipe redirections work out of the box.
- **macOS (Darwin)**: **Limited Compatibility.** POSIX syscalls are available, but process group signaling (`Setpgid`) and signal forwarding semantics can exhibit subtle differences compared to Linux.
- **Windows (Native)**: **Not Supported Natively.** Windows lacks native POSIX system calls used in `tish` (such as `syscall.SysProcAttr{Setpgid: true}`). Compiling or running natively on Windows `cmd.exe` or PowerShell will fail. Use **WSL** (Windows Subsystem for Linux) when running on Windows.

---

## 💡 Examples & Usage

### 1. Basic Commands & Built-ins
```bash
tish> pwd
tish> export GREETING=Hello
tish> echo $GREETING
tish> cd ..
```

### 2. Input/Output Redirection
```bash
# Truncate and write
tish> echo "hello world" > greeting.txt

# Append output
tish> echo "second line" >> greeting.txt

# Read input file
tish> cat < greeting.txt
```

### 3. Command Pipelines
```bash
tish> cat greeting.txt | grep hello | wc -l
```

### 4. Background Jobs & Process Control
```bash
# Run process in background
tish> sleep 10 &

# List active background job PIDs
tish> jobs

# Block until all background jobs finish
tish> wait
```

---

## 🔮 Future Enhancements & Roadmap

- **Cross-Platform Support (Linux, macOS, Windows)**:
  - Utilize Go **Build Tags** (`//go:build linux`, `//go:build windows`) in OS-specific source files to abstract system call differences (e.g. mapping `Setpgid: true` on Linux to `CreationFlags: CREATE_NEW_PROCESS_GROUP` on Windows).
  - Use `filepath` standard library functions to handle cross-platform path separators (`/` vs `\`).
- **Integration Tests for `executePipeline`**:
  - Add end-to-end integration tests in `main_test.go` verifying live command execution, pipe streaming, input/output file creation, and non-zero exit code propagation.
- **Native Line Editing**:
  - Integrate a native Go readline library (such as `golang.org/x/term` or `github.com/chzyer/readline`) to provide builtin arrow key history and cursor control without requiring external utilities like `rlwrap`.

