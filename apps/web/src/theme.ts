"use client";

import { createTheme } from "@mui/material/styles";

// Single source of truth for palette, typography, and shape. Components consume
// this via the sx prop / useTheme — never hardcode colors or fonts.
export const theme = createTheme({
  palette: {
    mode: "light",
    primary: { main: "#2563eb" },
    secondary: { main: "#7c3aed" },
    background: { default: "#f8fafc" },
  },
  typography: {
    fontFamily:
      'system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
    h1: { fontSize: "2rem", fontWeight: 700 },
    h2: { fontSize: "1.5rem", fontWeight: 600 },
  },
  shape: { borderRadius: 10 },
});
