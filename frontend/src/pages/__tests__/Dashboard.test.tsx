import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { BrowserRouter } from "react-router-dom";
import Dashboard from "../Dashboard";
import { AuthProvider } from "../../contexts/AuthContext";

// Mock the auth service
vi.mock("../../services/authService", () => ({
  default: {
    getCurrentUser: vi.fn().mockResolvedValue({
      id: "123",
      email: "test@example.com",
      created_at: "2025-01-01",
    }),
  },
}));

const MockedDashboard = () => (
  <BrowserRouter>
    <AuthProvider>
      <Dashboard />
    </AuthProvider>
  </BrowserRouter>
);

describe("Dashboard Page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders dashboard with authenticated layout", () => {
    render(<MockedDashboard />);

    expect(screen.getByText(/welcome to your dashboard/i)).toBeInTheDocument();
    expect(screen.getByText(/dashboard coming soon/i)).toBeInTheDocument();
  });

  it("renders logout button in header", () => {
    render(<MockedDashboard />);

    const logoutButton = screen.getByLabelText(/logout/i);
    expect(logoutButton).toBeInTheDocument();
  });
});
