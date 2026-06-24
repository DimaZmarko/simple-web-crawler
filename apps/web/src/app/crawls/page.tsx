import NextLink from "next/link";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { CrawlList } from "@/components/crawl-list";

export default function CrawlsPage() {
  return (
    <Container maxWidth="md" sx={{ py: { xs: 3, md: 6 } }}>
      <Stack spacing={3}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          spacing={2}
        >
          <Typography variant="h1">Crawls</Typography>
          <Button component={NextLink} href="/crawls/new" variant="contained">
            New crawl
          </Button>
        </Stack>
        <CrawlList />
      </Stack>
    </Container>
  );
}
