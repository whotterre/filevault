# FileVault 🔒

FileVault is a powerful local command-line interface (CLI) designed to simplify file management. It allows users to effortlessly **upload, list, retrieve, and delete files** using intuitive, command-style inputs. Built specifically for GDGoC Backend Student Club, FileVault provides a robust and easy-to-use solution for your file storage needs.

-----

## 🚀 Getting Started

### Week 1 - Building the CLI

This milestone focuses on laying the groundwork for FileVault. You'll be implementing the CLI using the **Command Pattern** for a clean, modular structure and efficient command routing. We'll also establish a foundational **data storage layer** and ensure a clear separation of core business logic into a dedicated **service layer**.

Commands for Week 1
- Upload a file - vault upload <filepath>
- List uploaded files - vault list 
- Display a help menu showing the available commands
- Delete an uploaded file - vault delete <fileId>

# Week 2: CLI II
Extension of CLI to  a multi-user file system with:
- User registration and login
- Session/token management
- Per-user file ownership
- Access control enforcement​
Using Redis for session management
PBKDF2 for password hashing (with SHA256 algo)
SQLite3 as the backing store for user identity data


## 📂 Project Structure

```
filevault/
├── cli/
│   ├── commands/             # Individual command implementations
│   └── plex.go               # Command routing and execution
├── go.mod                    # Go module file
├── go.sum                    # Go module checksums
├── main.go                   # Application entry point
├── README.md                 # You are here!
├── services/
│   └── file_service.go       # Core business logic for file operations
├── storage/
│   ├── metadata.json         # Database for file metadata
│   └── uploads/              # Directory for uploaded files
└── utils/                    # Utility functions
```

-----

## 📋 Metadata Database Schema

FileVault maintains a simple, yet effective, metadata database using a JSON file (`metadata.json`). Each entry stores crucial information about your uploaded files:

```json
{
  "file_id": "string",      // Unique identifier (UUID) for the file
  "filename": "string",     // Original name of the file
  "size": "string",         // Size of the file in kilobytes (e.g., "1024KB")
  "path": "string",         // Relative path to the stored file (e.g., "./uploads/notes.txt")
  "uploaded_at": "datetime" // Timestamp of when the file was uploaded (e.g., "2025-07-04T11:49:02Z")
}
```

-----

## 🤝 Credits

**System Design:** Obatula Fuad ([GitHub](https://github.com/Akinwalee))

-----
