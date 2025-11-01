import React, { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import {
  Box,
  TextField,
  Button,
  Typography,
  Paper,
  Link,
  InputAdornment,
  IconButton,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
} from "@mui/material";
import {
  Visibility,
  VisibilityOff,
  CheckCircle,
  RadioButtonUnchecked,
} from "@mui/icons-material";
import { useAuth } from "../contexts/AuthContext";
import ErrorAlert from "../components/ErrorAlert";
import {
  validateEmail,
  validatePasswordRequirements,
  PASSWORD_REQUIREMENTS,
  type PasswordRequirements,
} from "../utils/validation";

interface RegisterFormData {
  email: string;
  password: string;
}

/**
 * Register Page
 * User registration form with email and password
 */
const Register: React.FC = () => {
  const navigate = useNavigate();
  const { register: registerUser } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
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
  } = useForm<RegisterFormData>({
    defaultValues: {
      email: "",
      password: "",
    },
  });

  // Watch password field for real-time validation feedback
  const password = watch("password");

  React.useEffect(() => {
    if (password) {
      setPasswordReqs(validatePasswordRequirements(password));
    }
  }, [password]);

  const onSubmit = async (data: RegisterFormData) => {
    setLoading(true);
    setError(null);

    try {
      await registerUser(data.email, data.password);
      navigate("/dashboard");
    } catch (err: unknown) {
      console.error("Registration error:", err);
      const error = err as {
        response?: { data?: { error?: string; message?: string } };
        message?: string;
      };
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        "Registration failed. Please try again.";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  const allRequirementsMet =
    passwordReqs.minLength &&
    passwordReqs.hasUppercase &&
    passwordReqs.hasLowercase &&
    passwordReqs.hasNumber;

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
          Create Account
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
          align="center"
          mb={3}
        >
          Sign up to get started
        </Typography>

        {error && <ErrorAlert error={error} onClose={() => setError(null)} />}

        <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
          {/* Email Field */}
          <Controller
            name="email"
            control={control}
            rules={{
              required: "Email is required",
              validate: (value) =>
                validateEmail(value) || "Please enter a valid email address",
            }}
            render={({ field }) => (
              <TextField
                {...field}
                label="Email"
                type="email"
                fullWidth
                margin="normal"
                error={!!errors.email}
                helperText={errors.email?.message}
                disabled={loading}
                autoComplete="email"
                autoFocus
              />
            )}
          />

          {/* Password Field */}
          <Controller
            name="password"
            control={control}
            rules={{
              required: "Password is required",
              validate: () =>
                allRequirementsMet || "Password does not meet requirements",
            }}
            render={({ field }) => (
              <TextField
                {...field}
                label="Password"
                type={showPassword ? "text" : "password"}
                fullWidth
                margin="normal"
                error={!!errors.password}
                helperText={errors.password?.message}
                disabled={loading}
                autoComplete="new-password"
                InputProps={{
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        onClick={handleTogglePasswordVisibility}
                        edge="end"
                        aria-label="toggle password visibility"
                        disabled={loading}
                      >
                        {showPassword ? <VisibilityOff /> : <Visibility />}
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
              />
            )}
          />

          {/* Password Requirements */}
          {password && (
            <Box mt={1} mb={2}>
              <Typography variant="caption" color="text.secondary" gutterBottom>
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

          {/* Register Button */}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            size="large"
            disabled={loading || !allRequirementsMet}
            sx={{ mt: 2, mb: 2 }}
          >
            {loading ? <CircularProgress size={24} /> : "Register"}
          </Button>

          {/* Link to Login */}
          <Box textAlign="center">
            <Typography variant="body2">
              Already have an account?{" "}
              <Link component={RouterLink} to="/login" underline="hover">
                Log in
              </Link>
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Box>
  );
};

export default Register;
