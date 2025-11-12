package cmd

import (
	"fmt"
	"syscall"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage authentication (login, logout, register)",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account",
	Long:  "Authenticate with your email and password to access the portfolio management system",
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your account",
	Long:  "Clear authentication tokens and logout",
	RunE:  runLogout,
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create a new account",
	Long:  "Register a new user account to start managing your portfolios",
	RunE:  runRegister,
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  "Show information about the currently authenticated user",
	RunE:  runWhoami,
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(whoamiCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Prompt for email
	email, err := cli.ReadInput("Email")
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}

	// Prompt for password (hidden)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	// Create API client
	client := cli.NewClientFromConfig(config)

	// Login request
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	var loginResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := client.Request("POST", "/api/v1/auth/login", loginReq, &loginResp); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save tokens
	config.AccessToken = loginResp.AccessToken
	config.RefreshToken = loginResp.RefreshToken
	if err := cli.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	cli.PrintSuccess("Successfully logged in!")
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("User", loginResp.User.Name))
	fmt.Println(cli.RenderKeyValue("Email", loginResp.User.Email))

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := cli.ClearTokens(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	cli.PrintSuccess("Successfully logged out!")
	return nil
}

func runRegister(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Prompt for user details
	name, err := cli.ReadInput("Full Name")
	if err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}

	email, err := cli.ReadInput("Email")
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}

	// Prompt for password (hidden)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	// Confirm password
	fmt.Print("Confirm Password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	confirm := string(confirmBytes)

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	// Create API client
	client := cli.NewClientFromConfig(config)

	// Register request
	registerReq := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}

	var registerResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := client.Request("POST", "/api/v1/auth/register", registerReq, &registerResp); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Save tokens
	config.AccessToken = registerResp.AccessToken
	config.RefreshToken = registerResp.RefreshToken
	if err := cli.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	cli.PrintSuccess("Successfully registered and logged in!")
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("User", registerResp.User.Name))
	fmt.Println(cli.RenderKeyValue("Email", registerResp.User.Email))

	return nil
}

func runWhoami(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.AccessToken == "" {
		cli.PrintWarning("Not logged in. Use 'portfolios auth login' to authenticate.")
		return nil
	}

	client := cli.NewClientFromConfig(config)

	var userResp struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := client.Request("GET", "/api/v1/auth/me", nil, &userResp); err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	fmt.Println(cli.RenderSection("Current User"))
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("ID", fmt.Sprintf("%d", userResp.ID)))
	fmt.Println(cli.RenderKeyValue("Name", userResp.Name))
	fmt.Println(cli.RenderKeyValue("Email", userResp.Email))

	return nil
}
