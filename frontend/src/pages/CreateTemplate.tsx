/*  Валера снова здесь: докручиваю ширину и чиню динамический листинг.
    Нужны @mui/material, @mui/lab, @emotion/*, notistack – без них вода не пойдёт.
*/
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Button,
  Chip,
  Container,
  Grid,
  LinearProgress,
  Paper,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
  useMediaQuery
} from '@mui/material';
import { LoadingButton } from '@mui/lab';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DownloadIcon from '@mui/icons-material/Download';
import { useTheme } from '@mui/material/styles';
import { Controller, useForm } from 'react-hook-form';
import { useEffect, useRef, useState } from 'react';
import { enqueueSnackbar } from 'notistack';
import {
  AdvancedInput,
  CreateTemplateInput,
  DatabaseInput,
  DatabaseType,
  EndpointInput,
  ServiceProtocol,
  ServiceRole,
  TemplateStatus
} from '../types/graphql';
import { useTemplates } from '../hooks/useTemplates';

/* -------- формы -------- */
interface FormData {
  name: string;
  protocols: ServiceProtocol[];
  role: ServiceRole;
  databaseType: DatabaseType | null;
  ddl: string;
  generateSwaggerDocs: boolean;
}

export default function CreateTemplate() {
  const theme = useTheme();
  const isLgUp = useMediaQuery(theme.breakpoints.up('lg'));
  const progressRef = useRef<HTMLDivElement>(null);

  /* ------------ состояние ------------ */
  const [templateId, setTemplateId] = useState<string | null>(null);
  const [progressValue, setProgressValue] = useState(0);

  const {
    template,
    templateError,
    refetchTemplate,
    createTemplate
  } = useTemplates(templateId || undefined);

  /* ------------ форма ------------ */
  const {
    handleSubmit,
    control,
    setValue,
    watch,
    formState: { errors, isSubmitting }
  } = useForm<FormData>({
    defaultValues: {
      name: '',
      protocols: [],
      role: ServiceRole.SERVER,
      databaseType: null,
      ddl: '',
      generateSwaggerDocs: false
    }
  });

  const w = watch(); // все поля
  const showAdvanced =
    w.databaseType !== null || w.protocols.includes(ServiceProtocol.REST);

  /* ------------ пуллинг статуса ------------ */
  useEffect(() => {
    if (!templateId) return;
    refetchTemplate(templateId);
    const id = setInterval(() => refetchTemplate(templateId), 1500);
    return () => clearInterval(id);
  }, [templateId, refetchTemplate]);

  /* ------------ индикатор прогресса ------------ */
  useEffect(() => {
    if (!templateId) return;
    if (
      template?.status === TemplateStatus.COMPLETED ||
      template?.status === TemplateStatus.FAILED
    ) {
      setProgressValue(100);
      return;
    }
    const id = setInterval(
      () => setProgressValue(v => Math.min(v + 4, 92)),
      350
    );
    return () => clearInterval(id);
  }, [templateId, template?.status]);

  /* ------------ file-tree генератор ------------ */
  const generateFileTree = () => {
    const service = w.name || 'my-service';
    const hasGraphQL = w.protocols.includes(ServiceProtocol.GRAPHQL);
    const hasGRPC = w.protocols.includes(ServiceProtocol.GRPC);
    const hasREST = w.protocols.includes(ServiceProtocol.REST);
    const hasDB = w.databaseType !== null;

    const lines: string[] = [];
    const push = (lvl: number, s: string) =>
      lines.push(`${'  '.repeat(lvl)}${s}`);

    push(0, `${service}/`);
    /* API блок -------------------------------------------------- */
    push(1, '├── api/');
    if (hasGraphQL) {
      push(2, '├── graphql/');
      push(3, '└── users-posts-demo.graphql');
    }
    if (hasGRPC) {
      push(2, '├── grpc/');
      push(3, '└── users-posts-demo.proto');
    }
    if (hasREST) {
      push(2, '└── rest/');
      push(3, '└── users-posts-demo.yaml');
    }

    /* Build-infra ------------------------------------------------ */
    push(1, '├── build/');
    push(2, '└── config/');
    push(3, '└── config.yml');

    push(1, '├── cmd/');
    push(2, '└── main.go');

    push(1, '├── config/');
    push(2, '└── config.go');

    /* Internal --------------------------------------------------- */
    push(1, '├── internal/');
    push(2, '├── app/');
    push(3, '└── app.go');

    if (hasDB) {
      push(2, '├── database/');
      push(3, '├── default_repo/');
      push(4, '└── repository.go');
      push(3, '├── models/');
      push(4, '└── models.go');
      push(3, '├── gorm_repository.go');
      push(3, '└── implementation.go');
    }

    if (hasGraphQL) {
      push(2, '├── graphql/');
      push(3, '├── create_post.go');
      push(3, '└── create_user.go');
    }

    if (hasGRPC) {
      push(2, '├── grpc/');
      push(3, '├── create_post.go');
      push(3, '└── create_user.go');
    }

    push(2, '└── service/');
    push(3, '└── service.go');

    /* pkg -------------------------------------------------------- */
    push(1, '├── pkg/');
    push(2, '└── api/');
    if (hasGraphQL) push(3, '└── graphql/');
    if (hasGRPC) push(3, '└── grpc/');

    if (hasGraphQL) {
      push(1, '├── tools/');
      push(2, '└── tools.go');
      push(1, '├── gqlgen.yml');
    }

    /* root-files ------------------------------------------------- */
    push(1, '├── .gitignore');
    push(1, '├── go.mod');
    push(1, '├── go.sum');
    push(1, '├── Makefile');
    push(1, '├── README.md');
    push(1, '└── VERSION');

    return lines.join('\n');
  };

  /* ------------ загрузка DDL ------------ */
  const handleDdlUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const r = new FileReader();
    r.onload = ev => ev.target?.result && setValue('ddl', ev.target.result as string);
    r.readAsText(file);
  };

  /* ------------ submit ------------ */
  const onSubmit = async (data: FormData) => {
    const endpoints: EndpointInput[] | undefined = data.protocols.length
      ? data.protocols.map(p => ({ protocol: p, role: data.role }))
      : undefined;

    const database: DatabaseInput | undefined = data.databaseType
      ? { type: data.databaseType, ddl: data.ddl }
      : undefined;

    const advanced: AdvancedInput = { generateSwaggerDocs: data.generateSwaggerDocs };
    const input: CreateTemplateInput = { name: data.name, endpoints, database, advanced };

    try {
      const res = await createTemplate(input);
      if (res.success) {
        setTemplateId(res.templateId!);
        setTimeout(
          () => progressRef.current?.scrollIntoView({ behavior: 'smooth' }),
          100
        );
      } else {
        enqueueSnackbar(res.error ?? 'Failed to create template', { variant: 'error' });
      }
    } catch (e: any) {
      enqueueSnackbar(e.message ?? 'Unexpected error', { variant: 'error' });
    }
  };

  /* ------------ render ------------ */
  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      <Typography variant="h4" mb={3}>
        Generate service template
      </Typography>

      <Grid container spacing={isLgUp ? 6 : 4}>
        {/* ------ FORM ------- */}
        <Grid item xs={12} lg={7}>
          <Paper sx={{ p: 4, height: '100%' }}>
            <form onSubmit={handleSubmit(onSubmit)}>
              <Controller
                name="name"
                control={control}
                rules={{ required: 'Service name required' }}
                render={({ field }) => (
                  <TextField
                    {...field}
                    label="Service name"
                    fullWidth
                    error={!!errors.name}
                    helperText={errors.name?.message}
                  />
                )}
              />

              {/* Protocols */}
              <Box mt={4}>
                <Typography gutterBottom>Protocols</Typography>
                <Controller
                  name="protocols"
                  control={control}
                  render={({ field }) => (
                    <ToggleButtonGroup
                      value={field.value}
                      onChange={(_, v) => field.onChange(v)}
                      size="small"
                    >
                      <ToggleButton value={ServiceProtocol.GRPC}>gRPC</ToggleButton>
                      <ToggleButton value={ServiceProtocol.GRAPHQL}>GraphQL</ToggleButton>
                      <ToggleButton value={ServiceProtocol.REST}>REST</ToggleButton>
                    </ToggleButtonGroup>
                  )}
                />
              </Box>

              {/* DB */}
              <Box mt={4}>
                <Typography gutterBottom>Database (optional)</Typography>
                <Controller
                  name="databaseType"
                  control={control}
                  render={({ field }) => (
                    <ToggleButtonGroup
                      exclusive
                      value={field.value}
                      onChange={(_, v) => field.onChange(v)}
                      size="small"
                    >
                      <ToggleButton value={DatabaseType.POSTGRESQL}>PostgreSQL</ToggleButton>
                      <ToggleButton value={DatabaseType.MYSQL}>MySQL</ToggleButton>
                    </ToggleButtonGroup>
                  )}
                />
              </Box>

              {/* Advanced */}
              {showAdvanced && (
                <Accordion sx={{ mt: 4 }} TransitionProps={{ unmountOnExit: true }}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography>Advanced options</Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    {w.databaseType && (
                      <Stack spacing={2}>
                        <Button variant="outlined" component="label">
                          Upload DDL
                          <input hidden type="file" accept=".sql" onChange={handleDdlUpload} />
                        </Button>
                        <Controller
                          name="ddl"
                          control={control}
                          render={({ field }) => (
                            <TextField
                              {...field}
                              multiline
                              minRows={4}
                              label="DDL content"
                              placeholder="CREATE TABLE users (...);"
                            />
                          )}
                        />
                      </Stack>
                    )}

                    {w.protocols.includes(ServiceProtocol.REST) && (
                      <Controller
                        name="generateSwaggerDocs"
                        control={control}
                        render={({ field }) => (
                          <Chip
                            label="Generate Swagger docs"
                            color={field.value ? 'primary' : 'default'}
                            onClick={() => field.onChange(!field.value)}
                            variant={field.value ? 'filled' : 'outlined'}
                            sx={{ mt: 2 }}
                          />
                        )}
                      />
                    )}
                  </AccordionDetails>
                </Accordion>
              )}

              {/* actions */}
              <Stack direction="row" spacing={3} mt={5}>
                <LoadingButton
                  type="submit"
                  variant="contained"
                  loading={isSubmitting}
                  disabled={templateId && template?.status !== TemplateStatus.FAILED}
                >
                  Generate template
                </LoadingButton>
              </Stack>
            </form>
          </Paper>
        </Grid>

        {/* ------ PREVIEW ------- */}
        <Grid item xs={12} lg={5}>
          <Paper
            sx={{
              p: 4,
              height: isLgUp ? 'calc(100vh - 200px)' : 'auto',
              overflow: 'auto'
            }}
          >
            <Typography variant="h6" gutterBottom>
              Project structure preview
            </Typography>
            <pre style={{ fontFamily: 'monospace', fontSize: 14, whiteSpace: 'pre' }}>
              {generateFileTree()}
            </pre>
          </Paper>
        </Grid>
      </Grid>

      {/* ------ PROGRESS + DOWNLOAD ------- */}
      {templateId && (
        <Box mt={6} ref={progressRef}>
          <Paper sx={{ p: 4 }}>
            <Typography variant="h6" mb={2}>
              Build status
            </Typography>
            <LinearProgress
              variant="determinate"
              value={progressValue}
              color={template?.status === TemplateStatus.FAILED ? 'error' : 'primary'}
              sx={{ height: 10, borderRadius: 4 }}
            />

            <Box mt={2}>
              {template?.status && (
                <Chip
                  label={template.status}
                  color={
                    template.status === TemplateStatus.COMPLETED
                      ? 'success'
                      : template.status === TemplateStatus.FAILED
                      ? 'error'
                      : 'info'
                  }
                />
              )}
            </Box>

                        {template?.zipUrl && (
              <Button
                variant="contained"
                startIcon={<DownloadIcon />}
                sx={{ mt: 3 }}
                onClick={async () => {
                  try {
                    const response = await fetch(template.zipUrl);
                    if (!response.ok) {
                      throw new Error(`Ошибка загрузки: ${response.statusText}`);
                    }
                    const blob = await response.blob();
                    const filename = `${w.name || template.name || 'template'}.zip`;
                    const blobURL = window.URL.createObjectURL(blob);

                    const link = document.createElement('a');
                    link.href = blobURL;
                    link.download = filename;
                    document.body.appendChild(link);
                    link.click();
                    link.remove();
                    window.URL.revokeObjectURL(blobURL);
                  } catch (err) {
                    enqueueSnackbar('Ошибка при скачивании файла', { variant: 'error' });
                    console.error('Ошибка при скачивании ZIP:', err);
                  }
                }}
              >
                Download ZIP
              </Button>
            )}


            {(templateError || !template) && (
              <Typography color="error" mt={2}>
                {templateError?.message || 'Unknown error'}
              </Typography>
            )}
          </Paper>
        </Box>
      )}
    </Container>
  );
}
