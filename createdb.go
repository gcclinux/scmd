package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// SetupDatabase creates the scmd database and its tables from scratch.
// Run with: scmd --create-db
func SetupDatabase() {
	LoadEnv()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dstDB := os.Getenv("DB_NAME")
	dataTbl := os.Getenv("TB_NAME")
	embedDim := 384 // must match EMBEDDING_DIM in .env

	if dstDB == "" {
		dstDB = "scmd"
	}
	if dataTbl == "" {
		dataTbl = "data"
	}
	accessTbl := os.Getenv("ACCESS_TB")
	if accessTbl == "" {
		accessTbl = "access"
	}

	// -----------------------------------------------------------------------
	// Step 1 – Create the database (connect via postgres maintenance DB)
	// -----------------------------------------------------------------------
	fmt.Printf("\n=== Step 1: Create database '%s' ===\n", dstDB)

	maintConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, pass)
	maintDB, err := sql.Open("postgres", maintConn)
	if err != nil {
		log.Fatalf("Cannot connect to postgres: %v", err)
	}
	defer maintDB.Close()
	if err = maintDB.Ping(); err != nil {
		log.Fatalf("Cannot ping postgres maintenance DB: %v", err)
	}

	var exists bool
	maintDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dstDB).Scan(&exists)
	if exists {
		fmt.Printf("  Database '%s' already exists, skipping creation.\n", dstDB)
	} else {
		if _, err = maintDB.Exec(fmt.Sprintf(
			"CREATE DATABASE %s TEMPLATE template0 ENCODING 'UTF8'", dstDB)); err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		fmt.Printf("  Created database '%s'.\n", dstDB)
	}

	// -----------------------------------------------------------------------
	// Step 2 – Connect to the new database
	// -----------------------------------------------------------------------
	dstConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dstDB)
	conn, err := sql.Open("postgres", dstConn)
	if err != nil {
		log.Fatalf("Cannot open '%s': %v", dstDB, err)
	}
	defer conn.Close()
	if err = conn.Ping(); err != nil {
		log.Fatalf("Cannot ping '%s': %v", dstDB, err)
	}

	// -----------------------------------------------------------------------
	// Step 3 – Enable pgvector extension
	// -----------------------------------------------------------------------
	fmt.Println("\n=== Step 2: Enable pgvector extension ===")
	if _, err = conn.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		log.Fatalf("Failed to enable vector extension: %v", err)
	}
	fmt.Println("  pgvector ready.")

	// -----------------------------------------------------------------------
	// Step 4 – Create the main commands table
	// -----------------------------------------------------------------------
	fmt.Printf("\n=== Step 3: Create table '%s' ===\n", dataTbl)
	if _, err = conn.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id         SERIAL PRIMARY KEY,
			key        TEXT   NOT NULL,
			data       TEXT   NOT NULL,
			embedding  vector(%d),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, dataTbl, embedDim)); err != nil {
		log.Fatalf("Failed to create table '%s': %v", dataTbl, err)
	}
	fmt.Printf("  Table '%s' created.\n", dataTbl)

	// -----------------------------------------------------------------------
	// Step 5 – Create the access table
	// -----------------------------------------------------------------------
	fmt.Printf("\n=== Step 4: Create table '%s' ===\n", accessTbl)
	if _, err = conn.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id         SERIAL      PRIMARY KEY,
			email      VARCHAR(50) NOT NULL UNIQUE,
			api_key    VARCHAR(50) NOT NULL,
			created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
		)`, accessTbl)); err != nil {
		log.Fatalf("Failed to create '%s' table: %v", accessTbl, err)
	}
	fmt.Printf("  Table '%s' created.\n", accessTbl)

	// -----------------------------------------------------------------------
	// Done
	// -----------------------------------------------------------------------
	fmt.Println()
	fmt.Println("======================================================")
	fmt.Println("  Database setup complete!")
	fmt.Println("  Verify your .env matches:")
	fmt.Println()
	fmt.Printf("  DB_NAME=%s\n", dstDB)
	fmt.Printf("  TB_NAME=%s\n", dataTbl)
	fmt.Printf("  ACCESS_TB=%s\n", accessTbl)
	fmt.Println()
	fmt.Println("  Run --generate-embeddings after adding data.")
	fmt.Println("======================================================")
}
