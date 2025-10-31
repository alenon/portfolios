import React from 'react';
import { Box, Typography, Paper } from '@mui/material';
import AuthenticatedLayout from '../components/Layout/AuthenticatedLayout';
import { useAuth } from '../contexts/AuthContext';

/**
 * Dashboard Page
 * Main authenticated page (placeholder for future portfolio CRUD feature)
 */
const Dashboard: React.FC = () => {
  const { user } = useAuth();

  return (
    <AuthenticatedLayout>
      <Box>
        <Typography variant="h4" component="h1" gutterBottom>
          Welcome to Your Dashboard
        </Typography>

        <Paper sx={{ p: 3, mt: 3 }}>
          <Typography variant="h6" gutterBottom>
            Hello, {user?.email}!
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
            Dashboard coming soon...
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            This page will be expanded with portfolio management features in the future.
          </Typography>
        </Paper>
      </Box>
    </AuthenticatedLayout>
  );
};

export default Dashboard;
