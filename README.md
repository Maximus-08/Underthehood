# Tiny Shell (`tish`)

`tish` (Tiny Shell) is a lightweight, interactive shell written in Go from scratch. It is designed to demonstrate key systems programming concepts, including process lifecycle management (`fork`/`exec`), standard I/O redirection, and command pipelines.

---

## 🚀 Features

- **Interactive REPL**: A simple command loop with a custom prompt (`tish>`).
- **External Commands**: Executes standard system binaries available in the system path.
- **Built-in Commands**:
  - `cd <directory>`: Changes the current working directory.
  - `exit`: Gracefully exits the shell.
- **I/O Redirection**:
  - `< file`: Redirects standard input from a file.
  - `> file`: Redirects standard output to a file (truncating the file first).
  - `>> file`: Redirects standard output to a file (appending to the file).
- **Pipelines (`|`)**: Allows chaining multiple commands together via UNIX pipes, passing the output of one command as input to the next (e.g., `cat out.txt | grep appended`).

---

## 🛠️ Getting Started

### Prerequisites

- [Go](https://go.dev/) (version 1.22.2 or higher recommended)

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

## 💡 Examples & Usage

### 1. Basic Command Execution
```bash
tish> ls -la
tish> pwd
```

### 2. Input/Output Redirection
```bash
# Write output to a new file
tish> echo "hello world" > greeting.txt

# Append output to an existing file
tish> echo "new line" >> greeting.txt

# Read input from a file
tish> cat < greeting.txt
```

### 3. Command Pipelines
```bash
tish> cat greeting.txt | grep hello | wc -l
```

### 4. Built-in Directory Navigation
```bash
tish> cd ..
tish> pwd
```
