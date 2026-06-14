import Container from "@mui/material/Container";
import { CrawlDetail } from "@/components/crawl-detail";

export default async function CrawlDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <Container maxWidth="sm" sx={{ py: { xs: 3, md: 6 } }}>
      <CrawlDetail id={id} />
    </Container>
  );
}
