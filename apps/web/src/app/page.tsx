import NextLink from "next/link";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { ApiStatus } from "@/components/api-status";

export default function Home() {
  return (
    <Container maxWidth="md" sx={{ py: { xs: 4, md: 8 } }}>
      <Stack spacing={3} alignItems="flex-start">
        <Typography variant="h1">Simple Web Crawler</Typography>
        <ApiStatus />
        <Stack direction="row" spacing={2}>
          <Button component={NextLink} href="/crawls" variant="contained">
            View crawls
          </Button>
          <Button component={NextLink} href="/crawls/new" variant="outlined">
            New crawl
          </Button>
        </Stack>
      </Stack>
    </Container>
  );
}
