"use client";

import { useState } from "react";
import NextLink from "next/link";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "@mui/material/Link";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import { useCrawlList } from "@/api/crawls";

function formatDate(value: string): string {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString();
}

export function CrawlList() {
  // Holds the cursor for the page currently being viewed. The "Load more"
  // control advances it to the previous response's nextCursor.
  const [cursor, setCursor] = useState<string | undefined>(undefined);

  const { data, isPending, isError, error, isFetching } = useCrawlList(cursor);

  if (isPending) {
    return (
      <Stack alignItems="center" sx={{ py: 6 }}>
        <CircularProgress />
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

  if (data.items.length === 0) {
    return (
      <Stack spacing={2} alignItems="flex-start" sx={{ py: 4 }}>
        <Typography color="text.secondary">No crawls yet.</Typography>
        <Button component={NextLink} href="/crawls/new" variant="contained">
          Submit your first crawl
        </Button>
      </Stack>
    );
  }

  return (
    <Stack spacing={2}>
      <TableContainer component={Paper} variant="outlined">
        <Table aria-label="crawls">
          <TableHead>
            <TableRow>
              <TableCell>Seed URL</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Created</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.items.map((crawl) => (
              <TableRow key={crawl.id} hover>
                <TableCell>
                  <Link
                    component={NextLink}
                    href={`/crawls/${crawl.id}`}
                    underline="hover"
                  >
                    {crawl.seedUrl}
                  </Link>
                </TableCell>
                <TableCell>
                  <Chip
                    label={crawl.status}
                    size="small"
                    color="default"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell>{formatDate(crawl.createdAt)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Box>
        <Button
          variant="outlined"
          disabled={!data.nextCursor || isFetching}
          onClick={() => {
            if (data.nextCursor) {
              setCursor(data.nextCursor);
            }
          }}
        >
          {isFetching ? "Loading…" : data.nextCursor ? "Load more" : "No more"}
        </Button>
      </Box>
    </Stack>
  );
}
