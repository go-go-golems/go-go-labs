### Go Architecture for Rewriting `01-inspect-downloads-folder.sh` in Go

The task of rewriting the `01-inspect-downloads-folder.sh` script into a Go application involves several key components. The Go application will need to handle file system operations, execute external commands, and manage concurrency efficiently. Below is a proposed architecture for the Go application:

#### 1. **Main Package and Entry Point**
- **File:** `main.go`
- **Description:** This file will contain the `main()` function, which serves as the entry point of the application. It will handle command-line arguments and kick off the main processing logic.

#### 2. **Command-Line Interface**
- **Package:** `cli`
- **Description:** Utilize the `cobra` library to handle command-line arguments. This package will define commands, flags, and their descriptions. Flags will include options like `verbose` and `sample-per-dir`.

#### 3. **Configuration Management**
- **Package:** `config`
- **Description:** This package will parse and hold configuration settings derived from command-line arguments. It will provide an easy access point for other parts of the application to retrieve configuration settings.

#### 4. **File Analysis Logic**
- **Package:** `analysis`
- **Description:** Core logic for analyzing the downloads directory. This package will be responsible for:
  - **Directory Traversal:** Recursively traversing the file system starting from the downloads directory.
  - **File Metadata Collection:** Collecting metadata such as file size, modification time, and type.
  - **External Tool Invocation:** Running tools like `Magika`, `ExifTool`, and `jdupes` and parsing their output.
  - **Data Aggregation:** Aggregating data for reporting, such as counting file types and calculating total sizes.

#### 5. **Utility Functions**
- **Package:** `utils`
- **Description:** Contains helper functions used across the application. This could include:
  - **Human-Readable File Sizes:** Converting byte counts to a human-readable format.
  - **Error Handling:** Custom error handling utilities to wrap errors with context.

#### 6. **Logging and Output**
- **Package:** `logger`
- **Description:** Handles all logging operations. It will support different levels of logging (debug, info, error) and manage output to both the console and a debug log file.

#### 7. **Concurrency Management**
- **Package:** `concurrency`
- **Description:** Manages concurrent operations, particularly when analyzing files in large directories. This could use `sync` packages or `errgroup` for managing multiple goroutines.

#### 8. **External Command Wrapper**
- **Package:** `cmdwrapper`
- **Description:** Wraps external command execution, making it easier to run tools like `Magika`, `ExifTool`, and `jdupes` from Go code. This package will handle the intricacies of command execution and output collection.

#### 9. **Output Generation**
- **Package:** `output`
- **Description:** Manages the creation and formatting of the final analysis report. This includes writing to an output file and formatting the data into a readable format.

### High-Level Flow
1. **Parse Command-Line Arguments:** Use `cobra` to parse input options.
2. **Set Up Configuration:** Based on the parsed arguments, set up the global configuration.
3. **Initialize Logger:** Set up logging based on the verbosity level specified.
4. **Run Analysis:** Traverse the downloads directory, collect data, and use external tools as configured.
5. **Generate Report:** Aggregate collected data and generate a final report.
6. **Handle Errors and Output:** Throughout the process, handle errors gracefully and log necessary information.

This architecture aims to be modular, making each component responsible for a single aspect of the application, which aligns well with Go's design philosophy of simplicity and efficiency.
