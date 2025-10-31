import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Register from '../Register';
import { AuthProvider } from '../../contexts/AuthContext';

// Mock the auth service
vi.mock('../../services/authService', () => ({
  default: {
    register: vi.fn(),
    getCurrentUser: vi.fn(),
  },
}));

const MockedRegister = () => (
  <BrowserRouter>
    <AuthProvider>
      <Register />
    </AuthProvider>
  </BrowserRouter>
);

describe('Register Page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders registration form with all required fields', () => {
    render(<MockedRegister />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /register/i })).toBeInTheDocument();
  });

  it('displays password requirements when typing password', async () => {
    render(<MockedRegister />);

    const passwordInput = screen.getByLabelText(/^password$/i);

    fireEvent.change(passwordInput, { target: { value: 'Test' } });

    await waitFor(() => {
      expect(screen.getByText(/at least 8 characters/i)).toBeInTheDocument();
      expect(screen.getByText(/one uppercase letter/i)).toBeInTheDocument();
      expect(screen.getByText(/one lowercase letter/i)).toBeInTheDocument();
      expect(screen.getByText(/one number/i)).toBeInTheDocument();
    });
  });

  it('enables register button when all password requirements are met', async () => {
    render(<MockedRegister />);

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const registerButton = screen.getByRole('button', { name: /register/i });

    // Initially disabled
    expect(registerButton).toBeDisabled();

    // Fill in valid data
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'TestPass123' } });

    // Should be enabled
    await waitFor(() => {
      expect(registerButton).not.toBeDisabled();
    });
  });
});
