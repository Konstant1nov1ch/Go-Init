export interface Template {
  id: string;
  name: string;
  status: TemplateStatus;
  createdAt: string;
  updatedAt: string;
  endpoints?: EndpointConfig[];
  database?: DatabaseConfig;
  docker?: DockerConfig;
  advanced?: AdvancedConfig;
  error?: string;
  zipUrl?: string;
  version?: string;
}

export enum TemplateStatus {
  PENDING = 'PENDING',
  GENERATING = 'GENERATING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED'
}

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

export interface Framework {
  id: string;
  name: string;
  version: string;
  language: string;
}

export interface Database {
  id: string;
  name: string;
  type: string;
}

export interface Feature {
  id: string;
  name: string;
  description: string;
  compatible_frameworks: string[];
}

export interface TemplateStructure {
  folders: Folder[];
  files: File[];
}

export interface Folder {
  name: string;
  path: string;
}

export interface File {
  name: string;
  path: string;
  content?: string;
}

export interface CreateTemplateInput {
  name: string;
  endpoints?: EndpointInput[];
  database?: DatabaseInput;
  docker?: DockerInput;
  advanced?: AdvancedInput;
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

export interface UpdateTemplateInput {
  name?: string;
  endpoints?: EndpointInput[];
  database?: DatabaseInput;
  docker?: DockerInput;
  advanced?: AdvancedInput;
} 