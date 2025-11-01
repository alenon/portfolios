import React, { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import {
  Box,
  TextField,
  Button,
  Typography,
  Paper,
  InputAdornment,
  IconButton,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
  Alert,
} from "@mui/material";
import {
  Visibility,
  VisibilityOff,
  CheckCircle,
  RadioButtonUnchecked,
} from "@mui/icons-material";
import authService from "../services/authService";
import ErrorAlert from "../components/ErrorAlert";
import {
  validatePasswordRequirements,
  validatePasswordMatch,
  PASSWORD_REQUIREMENTS,
  type PasswordRequirements,
} from "../utils/validation";

interface ResetPasswordFormData {
  newPassword: string;
  confirmPassword: string;
}

/**
 * Reset Password Page
 * User can reset password using token from email
 */
const ResetPassword: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");

  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [passwordReqs, setPasswordReqs] = useState<PasswordRequirements>({
    minLength: false,
    hasUppercase: false,
    hasLowercase: false,
    hasNumber: false,
  });

  const {
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ResetPasswordFormData>({
    defaultValues: {
      newPassword: "",
      confirmPassword: "",
    },
  });

  // Watch password fields for real-time validation
  const newPassword = watch("newPassword");
  const confirmPassword = watch("confirmPassword");

  React.useEffect(() => {
    if (newPassword) {
      setPasswordReqs(validatePasswordRequirements(newPassword));
    }
  }, [newPassword]);

  // Check if token exists
  React.useEffect(() => {
    if (!token) {
      setError("Invalid reset link. Please request a new password reset.");
    }
  }, [token]);

  const onSubmit = async (data: ResetPasswordFormData) => {
    if (!token) {
      setError("Invalid reset link. Please request a new password reset.");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      await authService.resetPassword(token, data.newPassword);
      setSuccess(true);

      // Redirect to login after 3 seconds
      setTimeout(() => {
        navigate("/login");
      }, 3000);
    } catch (err: unknown) {
      console.error("Reset password error:", err);
      const error = err as {
        response?: { data?: { error?: string; message?: string } };
        message?: string;
      };
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        "Failed to reset password. Please try again or request a new reset link.";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const allRequirementsMet =
    passwordReqs.minLength &&
    passwordReqs.hasUppercase &&
    passwordReqs.hasLowercase &&
    passwordReqs.hasNumber;

  const passwordsMatch = validatePasswordMatch(newPassword, confirmPassword);

  const isFormValid = allRequirementsMet && passwordsMatch && !loading;

  return (
    <Box
      display="flex"
      justifyContent="center"
      alignItems="center"
      minHeight="100vh"
      sx={{
        backgroundColor: "background.default",
        p: 2,
      }}
    >
      <Paper
        elevation={3}
        sx={{
          p: 4,
          width: "100%",
          maxWidth: 450,
        }}
      >
        <Typography variant="h4" component="h1" align="center" gutterBottom>
          Reset Password
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
          align="center"
          mb={3}
        >
          Enter your new password below
        </Typography>

        {error && <ErrorAlert error={error} onClose={() => setError(null)} />}

        {success && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Password reset successfully! Redirecting to login...
          </Alert>
        )}

        {!success && token && (
          <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
            {/* New Password Field */}
            <Controller
              name="newPassword"
              control={control}
              rules={{
                required: "Password is required",
                validate: () =>
                  allRequirementsMet || "Password does not meet requirements",
              }}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="New Password"
                  type={showNewPassword ? "text" : "password"}
                  fullWidth
                  margin="normal"
                  error={!!errors.newPassword}
                  helperText={errors.newPassword?.message}
                  disabled={loading}
                  autoComplete="new-password"
                  autoFocus
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowNewPassword(!showNewPassword)}
                          edge="end"
                          aria-label="toggle password visibility"
                          disabled={loading}
                        >
                          {showNewPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
              )}
            />

            {/* Password Requirements */}
            {newPassword && (
              <Box mt={1} mb={2}>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  gutterBottom
                >
                  Password requirements:
                </Typography>
                <List dense disablePadding>
                  {PASSWORD_REQUIREMENTS.map((req) => {
                    const isMet = passwordReqs[req.key];
                    return (
                      <ListItem key={req.key} disablePadding sx={{ py: 0.25 }}>
                        <ListItemIcon sx={{ minWidth: 32 }}>
                          {isMet ? (
                            <CheckCircle fontSize="small" color="success" />
                          ) : (
                            <RadioButtonUnchecked
                              fontSize="small"
                              color="disabled"
                            />
                          )}
                        </ListItemIcon>
                        <ListItemText
                          primary={req.label}
                          primaryTypographyProps={{
                            variant: "caption",
                            color: isMet ? "success.main" : "text.secondary",
                          }}
                        />
                      </ListItem>
                    );
                  })}
                </List>
              </Box>
            )}

            {/* Confirm Password Field */}
            <Controller
              name="confirmPassword"
              control={control}
              rules={{
                required: "Please confirm your password",
                validate: () => passwordsMatch || "Passwords do not match",
              }}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Confirm Password"
                  type={showConfirmPassword ? "text" : "password"}
                  fullWidth
                  margin="normal"
                  error={
                    !!errors.confirmPassword ||
                    (confirmPassword && !passwordsMatch)
                  }
                  helperText={
                    errors.confirmPassword?.message ||
                    (confirmPassword && !passwordsMatch
                      ? "Passwords do not match"
                      : "")
                  }
                  disabled={loading}
                  autoComplete="new-password"
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() =>
                            setShowConfirmPassword(!showConfirmPassword)
                          }
                          edge="end"
                          aria-label="toggle password visibility"
                          disabled={loading}
                        >
                          {showConfirmPassword ? (
                            <VisibilityOff />
                          ) : (
                            <Visibility />
                          )}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
              )}
            />

            {/* Submit Button */}
            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={!isFormValid}
              sx={{ mt: 3, mb: 2 }}
            >
              {loading ? <CircularProgress size={24} /> : "Reset Password"}
            </Button>
          </Box>
        )}
      </Paper>
    </Box>
  );
};

export default ResetPassword;
