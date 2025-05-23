import { gql } from '@apollo/client';

export const GET_TEMPLATE = gql`
  query GetTemplate($id: ID!) {
    getTemplate(id: $id) {
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
        status
        error
      }
    }
  }
`;

export const GET_RECENT_TEMPLATES = gql`
  query GetRecentTemplates($limit: Int = 5) {
    getRecentTemplates(limit: $limit) {
      success
      message
      templates {
        id
        name
        status
        createdAt
        zipUrl
        error
      }
    }
  }
`; 