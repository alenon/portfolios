package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var (
	transactionType   string
	transactionBroker string
	dryRun            bool
)

var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"tx", "t"},
	Short:   "Manage transactions",
	Long:    "Add, import, and list portfolio transactions",
}

var transactionListCmd = &cobra.Command{
	Use:     "list <portfolio-id>",
	Aliases: []string{"ls", "l"},
	Short:   "List transactions",
	Long:    "Display all transactions for a specific portfolio",
	Args:    cobra.ExactArgs(1),
	RunE:    runTransactionList,
}

var transactionAddCmd = &cobra.Command{
	Use:   "add <portfolio-id>",
	Short: "Add a transaction",
	Long:  "Manually add a buy or sell transaction to a portfolio",
	Args:  cobra.ExactArgs(1),
	RunE:  runTransactionAdd,
}

var transactionImportCmd = &cobra.Command{
	Use:   "import <portfolio-id> <csv-file>",
	Short: "Import transactions from CSV",
	Long:  "Import transactions from a CSV file (supports multiple broker formats)",
	Args:  cobra.ExactArgs(2),
	RunE:  runTransactionImport,
}

var transactionDeleteCmd = &cobra.Command{
	Use:     "delete <portfolio-id> <transaction-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a transaction",
	Long:    "Delete a specific transaction from a portfolio",
	Args:    cobra.ExactArgs(2),
	RunE:    runTransactionDelete,
}

var transactionBatchListCmd = &cobra.Command{
	Use:     "batches <portfolio-id>",
	Aliases: []string{"batch"},
	Short:   "List import batches",
	Long:    "Display all import batches for a specific portfolio",
	Args:    cobra.ExactArgs(1),
	RunE:    runTransactionBatchList,
}

var transactionBatchDeleteCmd = &cobra.Command{
	Use:   "delete-batch <portfolio-id> <batch-id>",
	Short: "Delete an import batch",
	Long:  "Delete all transactions from a specific import batch",
	Args:  cobra.ExactArgs(2),
	RunE:  runTransactionBatchDelete,
}

func init() {
	transactionCmd.AddCommand(transactionListCmd)
	transactionCmd.AddCommand(transactionAddCmd)
	transactionCmd.AddCommand(transactionImportCmd)
	transactionCmd.AddCommand(transactionDeleteCmd)
	transactionCmd.AddCommand(transactionBatchListCmd)
	transactionCmd.AddCommand(transactionBatchDeleteCmd)

	// Add flags
	transactionAddCmd.Flags().StringVarP(&transactionType, "type", "t", "buy", "Transaction type (buy|sell)")
	transactionImportCmd.Flags().StringVarP(&transactionBroker, "broker", "b", "generic", "Broker format (generic|fidelity|schwab|tdameritrade|etrade|interactivebrokers|robinhood)")
	transactionImportCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate without importing")
}

func runTransactionList(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	var transactions []struct {
		ID              uint    `json:"id"`
		Type            string  `json:"type"`
		Symbol          string  `json:"symbol"`
		Quantity        float64 `json:"quantity"`
		Price           float64 `json:"price"`
		Commission      float64 `json:"commission"`
		TransactionDate string  `json:"transaction_date"`
		Notes           string  `json:"notes"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/"+portfolioID+"/transactions", nil, &transactions); err != nil {
		return err
	}

	if len(transactions) == 0 {
		cli.PrintInfo("No transactions found for this portfolio")
		return nil
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	headers := []string{"ID", "Type", "Symbol", "Quantity", "Price", "Commission", "Date", "Notes"}
	rows := make([][]string, len(transactions))

	for i, tx := range transactions {
		rows[i] = []string{
			fmt.Sprintf("%d", tx.ID),
			strings.ToUpper(tx.Type),
			tx.Symbol,
			fmt.Sprintf("%.4f", tx.Quantity),
			fmt.Sprintf("$%.2f", tx.Price),
			fmt.Sprintf("$%.2f", tx.Commission),
			tx.TransactionDate,
			truncate(tx.Notes, 30),
		}
	}

	return cli.Output(format, headers, rows, transactions)
}

func runTransactionAdd(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]

	// Prompt for transaction details
	txType, err := cli.ReadInput(fmt.Sprintf("Transaction Type [%s]", transactionType))
	if err != nil {
		return err
	}
	if txType == "" {
		txType = transactionType
	}

	symbol, err := cli.ReadInput("Symbol (e.g., AAPL)")
	if err != nil {
		return err
	}

	quantityStr, err := cli.ReadInput("Quantity")
	if err != nil {
		return err
	}
	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	priceStr, err := cli.ReadInput("Price per share")
	if err != nil {
		return err
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return fmt.Errorf("invalid price: %w", err)
	}

	commissionStr, err := cli.ReadInput("Commission (0 if none)")
	if err != nil {
		return err
	}
	commission := 0.0
	if commissionStr != "" {
		commission, err = strconv.ParseFloat(commissionStr, 64)
		if err != nil {
			return fmt.Errorf("invalid commission: %w", err)
		}
	}

	date, err := cli.ReadInput("Date (YYYY-MM-DD)")
	if err != nil {
		return err
	}

	notes, err := cli.ReadInput("Notes (optional)")
	if err != nil {
		return err
	}

	client := cli.NewClientFromConfig(config)

	createReq := map[string]interface{}{
		"type":             txType,
		"symbol":           strings.ToUpper(symbol),
		"quantity":         quantity,
		"price":            price,
		"commission":       commission,
		"transaction_date": date,
		"notes":            notes,
	}

	var transaction struct {
		ID              uint    `json:"id"`
		Type            string  `json:"type"`
		Symbol          string  `json:"symbol"`
		Quantity        float64 `json:"quantity"`
		Price           float64 `json:"price"`
		TransactionDate string  `json:"transaction_date"`
	}

	if err := client.Request("POST", "/api/v1/portfolios/"+portfolioID+"/transactions", createReq, &transaction); err != nil {
		return err
	}

	cli.PrintSuccess("Transaction added successfully!")
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("ID", fmt.Sprintf("%d", transaction.ID)))
	fmt.Println(cli.RenderKeyValue("Type", strings.ToUpper(transaction.Type)))
	fmt.Println(cli.RenderKeyValue("Symbol", transaction.Symbol))
	fmt.Println(cli.RenderKeyValue("Quantity", fmt.Sprintf("%.4f", transaction.Quantity)))
	fmt.Println(cli.RenderKeyValue("Price", fmt.Sprintf("$%.2f", transaction.Price)))
	fmt.Println(cli.RenderKeyValue("Total", fmt.Sprintf("$%.2f", transaction.Quantity*transaction.Price)))

	return nil
}

func runTransactionImport(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	csvFile := args[1]

	// Read CSV file
	fileData, err := os.ReadFile(csvFile)
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	client := cli.NewClientFromConfig(config)

	// Prepare additional fields
	additionalFields := map[string]string{
		"broker": transactionBroker,
	}
	if dryRun {
		additionalFields["dry_run"] = "true"
	}

	var importResp struct {
		Success       int      `json:"success"`
		Failed        int      `json:"failed"`
		Total         int      `json:"total"`
		Errors        []string `json:"errors"`
		ImportBatchID string   `json:"import_batch_id"`
		DryRun        bool     `json:"dry_run"`
	}

	if err := client.UploadFile(
		"/api/v1/portfolios/"+portfolioID+"/transactions/import/csv",
		"file",
		csvFile,
		fileData,
		additionalFields,
		&importResp,
	); err != nil {
		return err
	}

	if importResp.DryRun {
		cli.PrintInfo("Dry run completed - no transactions were imported")
	} else {
		cli.PrintSuccess("Import completed!")
	}

	fmt.Println()
	fmt.Println(cli.RenderKeyValue("Total", fmt.Sprintf("%d", importResp.Total)))
	fmt.Println(cli.RenderKeyValue("Success", fmt.Sprintf("%d", importResp.Success)))
	fmt.Println(cli.RenderKeyValue("Failed", fmt.Sprintf("%d", importResp.Failed)))

	if !importResp.DryRun && importResp.ImportBatchID != "" {
		fmt.Println(cli.RenderKeyValue("Batch ID", importResp.ImportBatchID))
	}

	if len(importResp.Errors) > 0 {
		fmt.Println()
		cli.PrintWarning("Errors encountered:")
		for _, errMsg := range importResp.Errors {
			fmt.Println("  - " + errMsg)
		}
	}

	return nil
}

func runTransactionDelete(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	transactionID := args[1]

	if !cli.Confirm(fmt.Sprintf("Are you sure you want to delete transaction %s?", transactionID)) {
		cli.PrintInfo("Deletion cancelled")
		return nil
	}

	client := cli.NewClientFromConfig(config)

	if err := client.Request("DELETE", "/api/v1/portfolios/"+portfolioID+"/transactions/"+transactionID, nil, nil); err != nil {
		return err
	}

	cli.PrintSuccess("Transaction deleted successfully!")
	return nil
}

func runTransactionBatchList(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	var batches []struct {
		ImportBatchID    string `json:"import_batch_id"`
		TransactionCount int    `json:"transaction_count"`
		FirstImportDate  string `json:"first_import_date"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/"+portfolioID+"/imports/batches", nil, &batches); err != nil {
		return err
	}

	if len(batches) == 0 {
		cli.PrintInfo("No import batches found for this portfolio")
		return nil
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	headers := []string{"Batch ID", "Transactions", "Import Date"}
	rows := make([][]string, len(batches))

	for i, b := range batches {
		rows[i] = []string{
			b.ImportBatchID,
			fmt.Sprintf("%d", b.TransactionCount),
			b.FirstImportDate,
		}
	}

	return cli.Output(format, headers, rows, batches)
}

func runTransactionBatchDelete(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	batchID := args[1]

	if !cli.Confirm(fmt.Sprintf("Are you sure you want to delete all transactions from batch %s?", batchID)) {
		cli.PrintInfo("Deletion cancelled")
		return nil
	}

	client := cli.NewClientFromConfig(config)

	if err := client.Request("DELETE", "/api/v1/portfolios/"+portfolioID+"/imports/batches/"+batchID, nil, nil); err != nil {
		return err
	}

	cli.PrintSuccess("Import batch deleted successfully!")
	return nil
}
