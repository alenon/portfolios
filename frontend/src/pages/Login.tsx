import React, { useState } from 'react';
import { useNavigate, useLocation, Link as RouterLink } from 'react-router-dom';
import { useForm, Controller } from 'react-hook-form';
import {
  Box,
  TextField,
  Button,
  Typography,
  Paper,
  Link,
  InputAdornment,
  IconButton,
  Checkbox,
  FormControlLabel,
  CircularProgress,
} from '@mui/material';
import {
  Visibility,
  VisibilityOff,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import ErrorAlert from '../components/ErrorAlert';
import { validateEmail } from '../utils/validation';

interface LoginFormData {
  email: string;
  password: string;
  rememberMe: boolean;
}

/**
 * Login Page
 * User login form with email, password, and remember me checkbox
 */
const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Get intended destination from location state or default to dashboard
  const from = (location.state as any)?.from?.pathname || '/dashboard';

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    defaultValues: {
      email: '',
      password: '',
      rememberMe: false,
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setLoading(true);
    setError(null);

    try {
      await login(data.email, data.password, data.rememberMe);
      // Redirect to intended destination
      navigate(from, { replace: true });
    } catch (err: any) {
      console.error('Login error:', err);
      const errorMessage =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        'Login failed. Please check your credentials.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  return (
    <Box
      display="flex"
      justifyContent="center"
      alignItems="center"
      minHeight="100vh"
      sx={{
        backgroundColor: 'background.default',
        p: 2,
      }}
    >
      <Paper
        elevation={3}
        sx={{
          p: 4,
          width: '100%',
          maxWidth: 450,
        }}
      >
        <Typography variant="h4" component="h1" align="center" gutterBottom>
          Welcome Back
        </Typography>
        <Typography variant="body2" color="text.secondary" align="center" mb={3}>
          Log in to your account
        </Typography>

        {error && <ErrorAlert error={error} onClose={() => setError(null)} />}

        <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
          {/* Email Field */}
          <Controller
            name="email"
            control={control}
            rules={{
              required: 'Email is required',
              validate: (value) => validateEmail(value) || 'Please enter a valid email address',
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
              required: 'Password is required',
            }}
            render={({ field }) => (
              <TextField
                {...field}
                label="Password"
                type={showPassword ? 'text' : 'password'}
                fullWidth
                margin="normal"
                error={!!errors.password}
                helperText={errors.password?.message}
                disabled={loading}
                autoComplete="current-password"
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

          {/* Remember Me and Forgot Password */}
          <Box
            display="flex"
            justifyContent="space-between"
            alignItems="center"
            mt={1}
            mb={2}
          >
            <Controller
              name="rememberMe"
              control={control}
              render={({ field }) => (
                <FormControlLabel
                  control={
                    <Checkbox
                      {...field}
                      checked={field.value}
                      disabled={loading}
                      color="primary"
                    />
                  }
                  label="Remember Me"
                />
              )}
            />

            <Link
              component={RouterLink}
              to="/forgot-password"
              variant="body2"
              underline="hover"
            >
              Forgot Password?
            </Link>
          </Box>

          {/* Login Button */}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            size="large"
            disabled={loading}
            sx={{ mt: 2, mb: 2 }}
          >
            {loading ? <CircularProgress size={24} /> : 'Log In'}
          </Button>

          {/* Link to Register */}
          <Box textAlign="center">
            <Typography variant="body2">
              Don't have an account?{' '}
              <Link component={RouterLink} to="/register" underline="hover">
                Sign up
              </Link>
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Box>
  );
};

export default Login;
