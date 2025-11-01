/**
 * Email validation regex
 * Validates standard email format
 */
export const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Password validation requirements
 */
export interface PasswordRequirements {
  minLength: boolean;
  hasUppercase: boolean;
  hasLowercase: boolean;
  hasNumber: boolean;
}

/**
 * Validate email format
 */
export const validateEmail = (email: string): boolean => {
  return EMAIL_REGEX.test(email);
};

/**
 * Get email validation error message
 */
export const getEmailError = (email: string): string | null => {
  if (!email) {
    return "Email is required";
  }

  if (!validateEmail(email)) {
    return "Please enter a valid email address";
  }

  return null;
};

/**
 * Validate password requirements
 * Returns object with boolean flags for each requirement
 */
export const validatePasswordRequirements = (
  password: string,
): PasswordRequirements => {
  return {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumber: /[0-9]/.test(password),
  };
};

/**
 * Check if password meets all requirements
 */
export const isPasswordValid = (password: string): boolean => {
  const requirements = validatePasswordRequirements(password);
  return (
    requirements.minLength &&
    requirements.hasUppercase &&
    requirements.hasLowercase &&
    requirements.hasNumber
  );
};

/**
 * Get password validation error message
 */
export const getPasswordError = (password: string): string | null => {
  if (!password) {
    return "Password is required";
  }

  if (!isPasswordValid(password)) {
    return "Password must be at least 8 characters and contain uppercase, lowercase, and number";
  }

  return null;
};

/**
 * Validate password match
 */
export const validatePasswordMatch = (
  password: string,
  confirmPassword: string,
): boolean => {
  return password === confirmPassword && password.length > 0;
};

/**
 * Get password match error message
 */
export const getPasswordMatchError = (
  password: string,
  confirmPassword: string,
): string | null => {
  if (!confirmPassword) {
    return "Please confirm your password";
  }

  if (!validatePasswordMatch(password, confirmPassword)) {
    return "Passwords do not match";
  }

  return null;
};

/**
 * Password requirement descriptions for UI display
 */
export const PASSWORD_REQUIREMENTS = [
  { key: "minLength", label: "At least 8 characters" },
  { key: "hasUppercase", label: "One uppercase letter" },
  { key: "hasLowercase", label: "One lowercase letter" },
  { key: "hasNumber", label: "One number" },
] as const;

/**
 * Generic field validation error messages
 */
export const VALIDATION_MESSAGES = {
  required: "This field is required",
  emailInvalid: "Please enter a valid email address",
  passwordInvalid: "Password does not meet requirements",
  passwordMismatch: "Passwords do not match",
  minLength: (min: number) => `Must be at least ${min} characters`,
  maxLength: (max: number) => `Must be no more than ${max} characters`,
} as const;
