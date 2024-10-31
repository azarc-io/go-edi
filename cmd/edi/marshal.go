package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/azarc-io/go-edi/pkg/edi"
	"github.com/azarc-io/go-edi/pkg/schemas"
	"github.com/spf13/cobra"
)

var marshalCmd = &cobra.Command{
	Use:   "marshal",
	Short: "Marshal JSON data to EDI format using a specified schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaFile, _ := cmd.Flags().GetString("schema")
		outputFile, _ := cmd.Flags().GetString("output")

		// Load schema
		schemaData, err := os.ReadFile(schemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}
		schema, err := schemas.LoadSchema(schemaData)
		if err != nil {
			return fmt.Errorf("failed to load schema: %w", err)
		}

		// Read JSON input data (from file or stdin)
		var inputData []byte
		if len(args) > 0 {
			inputData, err = os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}
		} else {
			inputData, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}

		// Unmarshal JSON input into a generic map
		var data map[string]any
		if err := json.Unmarshal(inputData, &data); err != nil {
			return fmt.Errorf("failed to unmarshal input JSON: %w", err)
		}

		// Marshal data to EDI format
		ediData, err := edi.Marshal(schema, data)
		if err != nil {
			return fmt.Errorf("failed to marshal data to EDI: %w", err)
		}

		// Output the EDI data (to file or stdout)
		if outputFile != "" {
			if err := os.WriteFile(outputFile, ediData, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
		} else {
			fmt.Println(string(ediData))
		}

		return nil
	},
}

func init() {
	marshalCmd.Flags().StringP("schema", "s", "", "Path to the schema file (required)")
	marshalCmd.Flags().StringP("output", "o", "", "Path to the output file (optional, defaults to stdout)")
	if err := cobra.MarkFlagRequired(marshalCmd.Flags(), "schema"); err != nil {
		panic(err)
	}
}
