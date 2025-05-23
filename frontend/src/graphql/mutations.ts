import { gql } from '@apollo/client';

// Создание нового шаблона
export const CREATE_TEMPLATE = gql`
  mutation CreateTemplate($input: CreateTemplateInput!) {
    createTemplate(input: $input) {
      success
      message
      template {
        id
        name
        endpoints {
          protocol
          role
        }
        database {
          type
          ddl
        }
        docker {
          registry
          imageName
        }
        advanced {
          enableAuthentication
          generateSwaggerDocs
        }
        createdAt
        zipUrl
        version
      }
    }
  }
`;

