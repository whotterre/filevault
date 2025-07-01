# FileVault 
A local command-line interface (CLI) that enables users to easily upload, list, retrieve, and delete files using intuitive command-style inputs.
## Week 1 - Building the CLI
This milestone initiates the  You will implement this CLI leveraging the Command Pattern for a clean, modular structure and routing, while establishing a foundational data storage layer and clearly separating core business logic into a dedicated service layer.
## Metadata database schema
We maintain a database of sorts for each file's metadata via a JSON file
```json
{
    "file_id": "string", // UUID
    "filename": "string",
    "size": "string", // In kilobytes
    "path": "string", // ./uploads/notes.txt"
    "uploaded_at": "datetime" // Iykyk
}
```

# Folder Structure
filevault/
├── cli
│   ├── commands
│   └── plex.go
├── go.mod
├── go.sum
├── main.go
├── README.md
├── services
│   └── file_service.go
├── storage
│   ├── metadata.json
│   └── uploads
└── utils


# Credits
System Design - Obatula Fuad [Github]("https://github.com/Akinwalee")
