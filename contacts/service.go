package contacts

import (
	"database/sql"
	"fmt"
	"log"
)

// GetContacts retrieves contacts with pagination
func GetContacts(limit, offset int) ([]Contact, error) {
	log.Printf("GetContacts: Retrieving contacts with limit %d and offset %d", limit, offset)

	rows, err := DB.Query("SELECT id, first_name, last_name, phone, address FROM contacts LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		log.Printf("GetContacts: Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Phone, &c.Address); err != nil {
			log.Printf("GetContacts: Error scanning row: %v", err)
			return nil, err
		}
		contacts = append(contacts, c)
	}
	log.Printf("GetContacts: Retrieved %d contacts", len(contacts))
	return contacts, nil
}

// SearchContacts searches for contacts based on a search term
func SearchContacts(term string) ([]Contact, error) {
	log.Printf("SearchContacts: Searching for contacts with term '%s'", term)

	query := `
		SELECT id, first_name, last_name, phone, address 
		FROM contacts 
		WHERE first_name ILIKE $1 
		OR last_name ILIKE $1 
		OR phone ILIKE $1 
		OR address ILIKE $1`

	rows, err := DB.Query(query, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Phone, &c.Address); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// AddContact adds a new contact to the database and returns the ID
func AddContact(contact Contact) (string, error) {
	log.Printf("AddContact: Adding contact %v", contact)
	var id string
	err := DB.QueryRow(
		"INSERT INTO contacts (first_name, last_name, phone, address) VALUES ($1, $2, $3, $4) RETURNING id",
		contact.FirstName, contact.LastName, contact.Phone, contact.Address,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// UpdateContact updates an existing contact based on non-nil fields
func UpdateContact(id string, updatedContact UpdateContactRequest) error {
	log.Printf("UpdateContact: Updating contact with ID %s", id)
	query := "UPDATE contacts SET"
	var args []interface{}
	argCount := 1

	// Check each field for nil and add to the query if not nil
	if updatedContact.FirstName != nil {
		query += fmt.Sprintf(" first_name = $%d,", argCount)
		args = append(args, *updatedContact.FirstName)
		argCount++
	}
	if updatedContact.LastName != nil {
		query += fmt.Sprintf(" last_name = $%d,", argCount)
		args = append(args, *updatedContact.LastName)
		argCount++
	}
	if updatedContact.Phone != nil {
		query += fmt.Sprintf(" phone = $%d,", argCount)
		args = append(args, *updatedContact.Phone)
		argCount++
	}
	if updatedContact.Address != nil {
		query += fmt.Sprintf(" address = $%d,", argCount)
		args = append(args, *updatedContact.Address)
		argCount++
	}

	// Check if there are fields to update
	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Remove trailing comma and append the WHERE clause
	query = query[:len(query)-1] // Remove the trailing comma
	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	// Execute the query
	result, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("item not found with such id: %s", id)
	}
	return nil
}

// DeleteContact removes a contact from the database
func DeleteContact(id string) error {
	log.Printf("DeleteContact: Deleting contact with ID %s", id)
	result, err := DB.Exec("DELETE FROM contacts WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetContactByID retrieves a contact by its ID
func GetContactByID(id string) (Contact, error) {
	log.Printf("GetContactByID: Retrieving contact with ID %s", id)
	var contact Contact
	err := DB.QueryRow(
		"SELECT id, first_name, last_name, phone, address FROM contacts WHERE id = $1",
		id,
	).Scan(&contact.ID, &contact.FirstName, &contact.LastName, &contact.Phone, &contact.Address)
	if err != nil {
		return contact, err
	}
	return contact, nil
}
