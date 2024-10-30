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

var unmarshalCmd = &cobra.Command{
	Use:   "unmarshal",
	Short: "Unmarshal EDI data to JSON using a specified schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaFile, _ := cmd.Flags().GetString("schema")
		inputFile, _ := cmd.Flags().GetString("input")

		// Load schema
		schemaData, err := os.ReadFile(schemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}
		schema, err := schemas.LoadSchema(schemaData)
		if err != nil {
			return fmt.Errorf("failed to load schema: %w", err)
		}

		// Read input data
		var inputData []byte
		if inputFile != "" {
			inputData, err = os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}
		} else {
			inputData, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}

		// Unmarshal EDI data
		output := make(map[string]any)
		if err := edi.Unmarshal(schema, inputData, &output); err != nil {
			return fmt.Errorf("failed to unmarshal EDI data: %w", err)
		}

		// Print output in JSON format
		d, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(d))
		return nil
	},
}

func init() {
	unmarshalCmd.Flags().StringP("schema", "s", "", "Path to the schema file (required)")
	unmarshalCmd.Flags().StringP("input", "i", "", "Path to the input data file (optional, defaults to stdin)")
	if err := cobra.MarkFlagRequired(unmarshalCmd.Flags(), "schema"); err != nil {
		panic(err)
	}
}
