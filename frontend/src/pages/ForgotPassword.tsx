import React, { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { useForm, Controller } from 'react-hook-form';
import {
  Box,
  TextField,
  Button,
  Typography,
  Paper,
  Link,
  CircularProgress,
  Alert,
} from '@mui/material';
import { ArrowBack } from '@mui/icons-material';
import authService from '../services/authService';
import ErrorAlert from '../components/ErrorAlert';
import { validateEmail } from '../utils/validation';

interface ForgotPasswordFormData {
  email: string;
}

/**
 * Forgot Password Page
 * User can request password reset email
 */
const ForgotPassword: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormData>({
    defaultValues: {
      email: '',
    },
  });

  const onSubmit = async (data: ForgotPasswordFormData) => {
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      await authService.forgotPassword(data.email);
      setSuccess(true);

      // Hide success message after 5 seconds
      setTimeout(() => {
        setSuccess(false);
      }, 5000);
    } catch (err: any) {
      console.error('Forgot password error:', err);
      const errorMessage =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        'Failed to send reset email. Please try again.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
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
          Forgot Password?
        </Typography>
        <Typography variant="body2" color="text.secondary" align="center" mb={3}>
          Enter your email and we'll send you a password reset link
        </Typography>

        {error && <ErrorAlert error={error} onClose={() => setError(null)} />}

        {success && (
          <Alert severity="success" sx={{ mb: 2 }}>
            If an account exists with this email, we've sent a password reset link
          </Alert>
        )}

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

          {/* Submit Button */}
          <Button
            type="submit"
            fullWidth
            variant="contained"
            size="large"
            disabled={loading}
            sx={{ mt: 3, mb: 2 }}
          >
            {loading ? <CircularProgress size={24} /> : 'Send Reset Link'}
          </Button>

          {/* Back to Login Link */}
          <Box textAlign="center">
            <Link
              component={RouterLink}
              to="/login"
              underline="hover"
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 0.5,
              }}
            >
              <ArrowBack fontSize="small" />
              Back to login
            </Link>
          </Box>
        </Box>
      </Paper>
    </Box>
  );
};

export default ForgotPassword;
