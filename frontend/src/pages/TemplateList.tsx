/*  Главный холл с карточками шаблонов. */
import {
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  Grid,
  Skeleton,
  Stack,
  TextField,
  Typography,
  Alert
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { Link } from 'react-router-dom';
import { useEffect, useState, useMemo } from 'react';
import { useLazyQuery } from '@apollo/client';
import { GET_RECENT_TEMPLATES } from '../graphql/queries';
import { ServiceTemplate, TemplateStatus } from '../types/graphql';
import { useSnackbar } from 'notistack';

export default function TemplateList() {
  const [templates, setTemplates] = useState<ServiceTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const { enqueueSnackbar } = useSnackbar();

  const [fetchTemplates] = useLazyQuery(GET_RECENT_TEMPLATES, {
    variables: { limit: 50 },
    onCompleted: data => {
      setTemplates(data.getRecentTemplates?.templates ?? []);
      setLoading(false);
    },
    onError: (error) => {
      console.error('Failed to load templates:', error);
      setError(error.message);
      setLoading(false);
      enqueueSnackbar('Не удалось подключиться к серверу. Проверьте соединение.', { 
        variant: 'error',
        autoHideDuration: 5000
      });
    }
  });

  useEffect(() => {
    try {
      fetchTemplates();
    } catch (err) {
      console.error('Error fetching templates:', err);
      setError(err instanceof Error ? err.message : 'Неизвестная ошибка при загрузке данных');
      setLoading(false);
    }
  }, [fetchTemplates]);

  /* filter */
  const filtered = useMemo(
    () =>
      templates.filter(t => t.name.toLowerCase().includes(search.toLowerCase())),
    [search, templates]
  );

  if (loading) {
    return (
      <Box p={3}>
        <Skeleton variant="rectangular" height={48} />
        <Skeleton sx={{ mt: 2 }} height={200} />
      </Box>
    );
  }

  return (
    <Box p={3}>
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          Ошибка соединения с сервером: {error}
        </Alert>
      )}

      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} mb={3}>
        <TextField
          label="Search templates"
          value={search}
          onChange={e => setSearch(e.target.value)}
          sx={{ flexGrow: 1 }}
        />
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          component={Link}
          to="/create"
        >
          Create new
        </Button>
      </Stack>

      {error ? (
        <Box sx={{ mt: 3, p: 3, textAlign: 'center' }}>
          <Typography variant="h6" color="error">
            Не удалось загрузить шаблоны с сервера
          </Typography>
          <Typography variant="body1" sx={{ mt: 2 }}>
            Сервер не доступен или произошла ошибка сети. Попробуйте позже.
          </Typography>
          <Button 
            variant="outlined" 
            sx={{ mt: 2 }} 
            onClick={() => {
              setLoading(true);
              setError(null);
              fetchTemplates();
            }}
          >
            Повторить попытку
          </Button>
        </Box>
      ) : filtered.length === 0 ? (
        <Typography>No templates found.</Typography>
      ) : (
        <Grid container spacing={3}>
          {filtered.map(t => (
            <Grid key={t.id} item xs={12} sm={6} md={4}>
              <Card>
                <CardActionArea component={Link} to={`/template/${t.id}`}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      {t.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {new Date(t.createdAt).toLocaleString()}
                    </Typography>
                    <Chip
                      sx={{ mt: 1 }}
                      label={t.status}
                      color={
                        t.status === TemplateStatus.COMPLETED
                          ? 'success'
                          : t.status === TemplateStatus.FAILED
                            ? 'error'
                            : 'info'
                      }
                      size="small"
                    />
                  </CardContent>
                </CardActionArea>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}
