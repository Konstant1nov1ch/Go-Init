/*  Детали одного шаблона — открываем люк и смотрим, течёт ли. */
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  IconButton,
  Paper,
  Skeleton,
  Stack,
  Typography
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import RefreshIcon from '@mui/icons-material/Refresh';
import DownloadIcon from '@mui/icons-material/Download';

import { Link, useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useLazyQuery } from '@apollo/client';
import { GET_TEMPLATE } from '../graphql/queries';
import { ServiceTemplate, TemplateStatus } from '../types/graphql';
import { enqueueSnackbar } from 'notistack';
import { formatRelative } from 'date-fns';

type State = { template: ServiceTemplate | null; loading: boolean };

export default function TemplateDetails() {
  const { id } = useParams<{ id: string }>();
  const [state, setState] = useState<State>({ template: null, loading: true });

  const [getTemplate, { loading: gqlLoading }] = useLazyQuery(GET_TEMPLATE, {
    variables: { id },
    fetchPolicy: 'network-only',
    onCompleted: data => {
      if (data.getTemplate?.success) {
        setState({ template: data.getTemplate.template, loading: false });
      } else {
        enqueueSnackbar(data.getTemplate?.message ?? 'Failed to load template', { variant: 'error' });
      }
    },
    onError: e => enqueueSnackbar(e.message, { variant: 'error' })
  });

  useEffect(() => {
    if (id) getTemplate();
  }, [id, getTemplate]);

  const { template, loading } = state;

  if (loading || gqlLoading) {
    return (
      <Box p={4}>
        <Skeleton variant="rectangular" height={48} />
        <Skeleton sx={{ mt: 2 }} height={200} />
      </Box>
    );
  }

  if (!template) {
    return (
      <Box p={4}>
        <Typography variant="h6" color="error" gutterBottom>
          Template not found
        </Typography>
        <Button variant="contained" component={Link} to="/" startIcon={<ArrowBackIcon />}>
          Back
        </Button>
      </Box>
    );
  }

  return (
    <Box p={3}>
      {/* sticky toolbar */}
      <Stack direction="row" alignItems="center" spacing={1} mb={3}>
        <IconButton component={Link} to="/" aria-label="back">
          <ArrowBackIcon />
        </IconButton>
        <Typography variant="h5" flexGrow={1}>
          {template.name}
        </Typography>
        <IconButton onClick={() => getTemplate()} aria-label="refresh">
          <RefreshIcon />
        </IconButton>
      </Stack>

      <Paper elevation={2} sx={{ p: 3 }}>
        <Stack spacing={1}>
          <Typography>
            <strong>Created:</strong> {formatRelative(new Date(template.createdAt), new Date())}
          </Typography>
          {template.version && (
            <Typography>
              <strong>Version:</strong> {template.version}
            </Typography>
          )}
          <Chip
            label={template.status}
            color={
              template.status === TemplateStatus.COMPLETED
                ? 'success'
                : template.status === TemplateStatus.FAILED
                  ? 'error'
                  : 'info'
            }
            sx={{ width: 'fit-content' }}
          />
        </Stack>

        <Box mt={3}>
          {template.endpoints?.length ? (
            <Typography>
              <strong>Protocol:</strong> {template.endpoints[0].protocol}{' '}
              {template.endpoints[0].role && `(${template.endpoints[0].role})`}
            </Typography>
          ) : null}

          {template.database && (
            <Typography>
              <strong>Database:</strong> {template.database.type}
            </Typography>
          )}

          {template.docker && (
            <Typography>
              <strong>Docker:</strong> {template.docker.registry ?? ''}/{template.docker.imageName}
            </Typography>
          )}
        </Box>

        {/* download */}
        <Box mt={4}>
          {template.zipUrl ? (
            <Button
              variant="contained"
              startIcon={<DownloadIcon />}
              href={template.zipUrl}
              download={`${template.name}.zip`}
            >
              Download template
            </Button>
          ) : template.status !== TemplateStatus.COMPLETED ? (
            <Stack direction="row" spacing={2} alignItems="center">
              <CircularProgress size={20} />
              <Typography>Building, come back later…</Typography>
            </Stack>
          ) : (
            <Typography color="error">No download link</Typography>
          )}
        </Box>
      </Paper>
    </Box>
  );
}
