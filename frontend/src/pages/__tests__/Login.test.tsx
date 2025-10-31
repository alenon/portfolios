import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Login from '../Login';
import { AuthProvider } from '../../contexts/AuthContext';

// Mock the auth service
vi.mock('../../services/authService', () => ({
  default: {
    login: vi.fn(),
    getCurrentUser: vi.fn(),
  },
}));

const MockedLogin = () => (
  <BrowserRouter>
    <AuthProvider>
      <Login />
    </AuthProvider>
  </BrowserRouter>
);

describe('Login Page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form with all required fields', () => {
    render(<MockedLogin />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    // Use more specific selector for password field
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: /remember me/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
  });

  it('shows validation error when submitting empty form', async () => {
    render(<MockedLogin />);

    const submitButton = screen.getByRole('button', { name: /log in/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/email is required/i)).toBeInTheDocument();
    });
  });

  it('toggles password visibility when clicking the eye icon', () => {
    render(<MockedLogin />);

    const passwordInput = screen.getByLabelText('Password') as HTMLInputElement;
    const toggleButton = screen.getByLabelText(/toggle password visibility/i);

    expect(passwordInput.type).toBe('password');

    fireEvent.click(toggleButton);
    expect(passwordInput.type).toBe('text');

    fireEvent.click(toggleButton);
    expect(passwordInput.type).toBe('password');
  });

  it('contains link to register page', () => {
    render(<MockedLogin />);

    const registerLink = screen.getByText(/sign up/i);
    expect(registerLink).toBeInTheDocument();
    expect(registerLink.closest('a')).toHaveAttribute('href', '/register');
  });

  it('contains link to forgot password page', () => {
    render(<MockedLogin />);

    const forgotPasswordLink = screen.getByText(/forgot password/i);
    expect(forgotPasswordLink).toBeInTheDocument();
    expect(forgotPasswordLink.closest('a')).toHaveAttribute('href', '/forgot-password');
  });
});
