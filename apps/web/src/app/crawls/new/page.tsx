import Container from "@mui/material/Container";
import { CrawlForm } from "@/components/crawl-form";

export default function NewCrawlPage() {
  return (
    <Container maxWidth="sm" sx={{ py: { xs: 3, md: 6 } }}>
      <CrawlForm />
    </Container>
  );
}
