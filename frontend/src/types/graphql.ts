export enum ServiceProtocol {
  GRPC = 'GRPC',
  REST = 'REST',
  GRAPHQL = 'GRAPHQL'
}

export enum ServiceRole {
  CLIENT = 'CLIENT',
  SERVER = 'SERVER'
}

export enum DatabaseType {
  POSTGRESQL = 'POSTGRESQL',
  MYSQL = 'MYSQL',
  NONE = 'NONE'
}

export enum TemplateStatus {
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED'
}

export interface EndpointConfig {
  protocol: ServiceProtocol;
  role: ServiceRole;
}

export interface DatabaseConfig {
  type: DatabaseType;
  ddl?: string;
}

export interface DockerConfig {
  registry?: string;
  imageName: string;
}

export interface AdvancedConfig {
  enableAuthentication?: boolean;
  generateSwaggerDocs?: boolean;
}

export interface ServiceTemplate {
  id: string;
  name: string;
  endpoints?: EndpointConfig[];
  database?: DatabaseConfig;
  docker?: DockerConfig;
  advanced?: AdvancedConfig;
  createdAt: string;
  updatedAt?: string;
  zipUrl?: string;
  version?: string;
  status?: TemplateStatus;
  error?: string;
}

export interface CreateTemplateInput {
  name: string;
  endpoints?: EndpointInput[];
  database?: DatabaseInput;
  docker?: DockerInput;
  advanced?: AdvancedInput;
}

export interface UpdateTemplateInput {
  name?: string;
  endpoints?: EndpointInput[];
  database?: DatabaseInput;
  docker?: DockerInput;
  advanced?: AdvancedInput;
}

export interface TemplateResponse {
  success: boolean;
  message?: string;
  template?: ServiceTemplate;
}

export interface TemplatesResponse {
  success: boolean;
  message?: string;
  templates?: ServiceTemplate[];
}

export interface EndpointInput {
  protocol: ServiceProtocol;
  role: ServiceRole;
}

export interface DatabaseInput {
  type: DatabaseType;
  ddl?: string;
}

export interface DockerInput {
  registry?: string;
  imageName: string;
}

export interface AdvancedInput {
  enableAuthentication?: boolean;
  generateSwaggerDocs?: boolean;
} 