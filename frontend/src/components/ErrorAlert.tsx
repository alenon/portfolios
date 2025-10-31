import React from 'react';
import { Alert, AlertTitle } from '@mui/material';

interface ErrorAlertProps {
  error: string | Error | null;
  title?: string;
  onClose?: () => void;
}

/**
 * ErrorAlert Component
 * Displays error messages using MUI Alert component
 */
const ErrorAlert: React.FC<ErrorAlertProps> = ({ error, title = 'Error', onClose }) => {
  if (!error) return null;

  const errorMessage = typeof error === 'string' ? error : error.message;

  return (
    <Alert severity="error" onClose={onClose} sx={{ width: '100%', mb: 2 }}>
      {title && <AlertTitle>{title}</AlertTitle>}
      {errorMessage}
    </Alert>
  );
};

export default ErrorAlert;
