"use client";

import NextLink from "next/link";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import Divider from "@mui/material/Divider";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { useCrawl } from "@/api/crawls";

function formatDate(value: string): string {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString();
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <Stack
      direction={{ xs: "column", sm: "row" }}
      justifyContent="space-between"
      spacing={1}
    >
      <Typography color="text.secondary">{label}</Typography>
      <Typography sx={{ wordBreak: "break-all" }}>{value}</Typography>
    </Stack>
  );
}

export function CrawlDetail({ id }: { id: string }) {
  const { data, isPending, isError, error } = useCrawl(id);

  if (isPending) {
    return (
      <Stack alignItems="center" sx={{ py: 6 }}>
        <CircularProgress />
      </Stack>
    );
  }

  // A 404 resolves to null rather than rejecting — render a not-found state.
  if (data === null) {
    return (
      <Stack spacing={2} alignItems="flex-start">
        <Alert severity="warning" role="alert">
          Crawl not found.
        </Alert>
        <Button component={NextLink} href="/crawls" variant="outlined">
          Back to crawls
        </Button>
      </Stack>
    );
  }

  if (isError) {
    return (
      <Alert severity="error" role="alert">
        {error.message}
      </Alert>
    );
  }

  return (
    <Stack spacing={3}>
      <Stack
        direction="row"
        justifyContent="space-between"
        alignItems="center"
        spacing={2}
      >
        <Typography variant="h1" sx={{ wordBreak: "break-all" }}>
          {data.seedUrl}
        </Typography>
        <Chip label={data.status} color="primary" variant="outlined" />
      </Stack>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2} divider={<Divider flexItem />}>
            <Field label="ID" value={data.id} />
            <Field label="Seed URL" value={data.seedUrl} />
            <Field label="Max depth" value={String(data.maxDepth)} />
            <Field label="Max pages" value={String(data.maxPages)} />
            <Field label="Status" value={data.status} />
            <Field label="Created" value={formatDate(data.createdAt)} />
            <Field label="Updated" value={formatDate(data.updatedAt)} />
          </Stack>
        </CardContent>
      </Card>

      <Button
        component={NextLink}
        href="/crawls"
        variant="outlined"
        sx={{ alignSelf: "flex-start" }}
      >
        Back to crawls
      </Button>
    </Stack>
  );
}
