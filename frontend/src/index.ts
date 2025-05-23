// Главный barrel-файл: »экспортируем публичное API пакета«
import { TemplateStatus } from './types/graphql';
import type {
  ServiceTemplate,
  CreateTemplateInput,
  UpdateTemplateInput
} from './types/graphql';

import { useTemplates } from './hooks/useTemplates';

// при желании можно заэкспортить Logo или Theme дальше
export { TemplateStatus, useTemplates };
export type {
  ServiceTemplate,
  CreateTemplateInput,
  UpdateTemplateInput
};
