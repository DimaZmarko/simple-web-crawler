"use client";

import { useState, type FormEvent } from "react";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import {
  CrawlValidationError,
  useCreateCrawl,
  type CreateCrawlRequest,
} from "@/api/crawls";

type FieldErrors = Partial<Record<keyof CreateCrawlRequest, string>>;

function isAbsoluteHttpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

// Client-side validation mirrors the contract bounds (http(s) URL, depth >= 0,
// pages >= 1). The server is still authoritative — its 400 field errors are
// merged on top of these in the submit handler.
function validate(values: {
  seedUrl: string;
  maxDepth: string;
  maxPages: string;
}): FieldErrors {
  const errors: FieldErrors = {};

  if (!values.seedUrl.trim()) {
    errors.seedUrl = "Seed URL is required";
  } else if (!isAbsoluteHttpUrl(values.seedUrl.trim())) {
    errors.seedUrl = "Must be an absolute http or https URL";
  }

  const depth = Number(values.maxDepth);
  if (values.maxDepth === "" || !Number.isInteger(depth) || depth < 0) {
    errors.maxDepth = "Max depth must be an integer of 0 or more";
  }

  const pages = Number(values.maxPages);
  if (values.maxPages === "" || !Number.isInteger(pages) || pages < 1) {
    errors.maxPages = "Max pages must be an integer of 1 or more";
  } else if (pages > 10000) {
    errors.maxPages = "Max pages must be at most 10000";
  }

  return errors;
}

export function CrawlForm() {
  const [seedUrl, setSeedUrl] = useState("");
  const [maxDepth, setMaxDepth] = useState("2");
  const [maxPages, setMaxPages] = useState("500");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  const createCrawl = useCreateCrawl();

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const clientErrors = validate({ seedUrl, maxDepth, maxPages });
    if (Object.keys(clientErrors).length > 0) {
      setFieldErrors(clientErrors);
      return;
    }
    setFieldErrors({});

    const body: CreateCrawlRequest = {
      seedUrl: seedUrl.trim(),
      maxDepth: Number(maxDepth),
      maxPages: Number(maxPages),
    };

    createCrawl.mutate(body, {
      onError: (error) => {
        // Render server-side 400 field errors next to their inputs.
        if (error instanceof CrawlValidationError) {
          const serverErrors: FieldErrors = {};
          for (const fieldError of error.errors) {
            serverErrors[fieldError.field as keyof CreateCrawlRequest] =
              fieldError.message;
          }
          setFieldErrors(serverErrors);
        }
      },
    });
  };

  const generalError =
    createCrawl.isError && !(createCrawl.error instanceof CrawlValidationError)
      ? createCrawl.error.message
      : null;

  return (
    <Box component="form" onSubmit={handleSubmit} noValidate>
      <Stack spacing={3}>
        <Typography variant="h1">New crawl</Typography>

        {generalError && (
          <Alert severity="error" role="alert">
            {generalError}
          </Alert>
        )}
        {createCrawl.isSuccess && (
          <Alert severity="success" role="status">
            Crawl queued. Redirecting…
          </Alert>
        )}

        <TextField
          label="Seed URL"
          name="seedUrl"
          value={seedUrl}
          onChange={(event) => setSeedUrl(event.target.value)}
          placeholder="https://example.com"
          error={Boolean(fieldErrors.seedUrl)}
          helperText={fieldErrors.seedUrl ?? "Absolute http or https URL"}
          fullWidth
          required
        />

        <TextField
          label="Max depth"
          name="maxDepth"
          type="number"
          value={maxDepth}
          onChange={(event) => setMaxDepth(event.target.value)}
          error={Boolean(fieldErrors.maxDepth)}
          helperText={fieldErrors.maxDepth ?? "Link depth to follow (>= 0)"}
          inputProps={{ min: 0, step: 1 }}
          fullWidth
          required
        />

        <TextField
          label="Max pages"
          name="maxPages"
          type="number"
          value={maxPages}
          onChange={(event) => setMaxPages(event.target.value)}
          error={Boolean(fieldErrors.maxPages)}
          helperText={fieldErrors.maxPages ?? "Maximum pages to fetch (>= 1)"}
          inputProps={{ min: 1, step: 1 }}
          fullWidth
          required
        />

        <Button
          type="submit"
          variant="contained"
          size="large"
          disabled={createCrawl.isPending || createCrawl.isSuccess}
          sx={{ alignSelf: "flex-start" }}
        >
          {createCrawl.isPending ? "Submitting…" : "Submit crawl"}
        </Button>
      </Stack>
    </Box>
  );
}
