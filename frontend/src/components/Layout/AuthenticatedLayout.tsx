import React from "react";
import { useNavigate } from "react-router-dom";
import {
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Box,
  Container,
} from "@mui/material";
import {
  Logout as LogoutIcon,
  AccountCircle as AccountCircleIcon,
} from "@mui/icons-material";
import { useAuth } from "../../contexts/AuthContext";

interface AuthenticatedLayoutProps {
  children: React.ReactNode;
}

/**
 * AuthenticatedLayout Component
 * Provides header with user info and logout button for authenticated pages
 */
const AuthenticatedLayout: React.FC<AuthenticatedLayoutProps> = ({
  children,
}) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate("/login");
    } catch (error) {
      console.error("Logout error:", error);
      // Navigate to login even if logout fails
      navigate("/login");
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      {/* Header */}
      <AppBar position="static">
        <Toolbar>
          {/* App Logo/Name */}
          <Typography variant="h6" component="div" sx={{ flexGrow: 0, mr: 4 }}>
            Portfolios
          </Typography>

          {/* Spacer */}
          <Box sx={{ flexGrow: 1 }} />

          {/* User Email Display */}
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              gap: 1,
              mr: 2,
            }}
          >
            <AccountCircleIcon />
            <Typography
              variant="body1"
              sx={{ display: { xs: "none", sm: "block" } }}
            >
              {user?.email}
            </Typography>
          </Box>

          {/* Logout Button */}
          <IconButton
            color="inherit"
            onClick={handleLogout}
            aria-label="logout"
            title="Logout"
          >
            <LogoutIcon />
          </IconButton>
        </Toolbar>
      </AppBar>

      {/* Main Content */}
      <Container
        component="main"
        maxWidth="lg"
        sx={{
          flexGrow: 1,
          py: 4,
        }}
      >
        {children}
      </Container>

      {/* Footer (Optional) */}
      <Box
        component="footer"
        sx={{
          py: 2,
          px: 2,
          mt: "auto",
          backgroundColor: (theme) =>
            theme.palette.mode === "light"
              ? theme.palette.grey[200]
              : theme.palette.grey[800],
        }}
      >
        <Container maxWidth="lg">
          <Typography variant="body2" color="text.secondary" align="center">
            © {new Date().getFullYear()} Portfolios App. All rights reserved.
          </Typography>
        </Container>
      </Box>
    </Box>
  );
};

export default AuthenticatedLayout;
