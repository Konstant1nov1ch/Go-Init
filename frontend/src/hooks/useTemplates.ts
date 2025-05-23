import { useMutation, useQuery, ApolloError } from '@apollo/client';
import { 
  CREATE_TEMPLATE
} from '../graphql/mutations';
import { 
  GET_TEMPLATE
} from '../graphql/queries';
import { 
  ServiceTemplate as Template, 
  CreateTemplateInput,
  TemplateResponse
} from '../types/graphql';

interface UseTemplatesReturn {
  // Queries
  template: Template | undefined;
  templateLoading: boolean;
  templateError: ApolloError | undefined;
  
  // Mutations
  createTemplate: (input: CreateTemplateInput) => Promise<{ success: boolean; templateId?: string; error?: string }>;
  createTemplateLoading: boolean;
  
  // Refetch functions
  refetchTemplate: (id: string) => Promise<void>;
}

export function useTemplates(templateId?: string): UseTemplatesReturn {
  // Queries
  const { 
    data: templateData, 
    loading: templateLoading, 
    error: templateError,
    refetch: refetchTemplateQuery  
  } = useQuery<{ getTemplate: TemplateResponse }>(GET_TEMPLATE, { 
    variables: { id: templateId },
    skip: !templateId,
    fetchPolicy: 'network-only', // Force network request, don't use cache
    nextFetchPolicy: 'network-only', // Also use network-only for subsequent requests
    errorPolicy: 'all', // Handle errors in the component
  });
  
  // Mutations
  const [createTemplateMutation, { loading: createTemplateLoading }] = useMutation<{ createTemplate: TemplateResponse }>(CREATE_TEMPLATE);
  
  // Derived data
  const template = templateData?.getTemplate?.template;
  
  // Refetch functions
  const refetchTemplate = async (id: string): Promise<void> => {
    try {
      await refetchTemplateQuery({ 
        id,
        fetchPolicy: 'network-only', // Всегда запрашивать с сервера, игнорировать кэш
      });
    } catch (error) {
      console.error('Error refetching template:', error);
    }
  };
  
  // Mutation handlers
  const createTemplate = async (input: CreateTemplateInput): Promise<{ success: boolean; templateId?: string; error?: string }> => {
    try {
      const { data } = await createTemplateMutation({
        variables: { input }
      });
      
      if (data?.createTemplate?.success && data.createTemplate.template) {
        return {
          success: true,
          templateId: data.createTemplate.template.id
        };
      } else {
        return {
          success: false,
          error: data?.createTemplate?.message || 'Unknown error'
        };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      };
    }
  };
  
  return {
    // Queries
    template,
    templateLoading,
    templateError,
    
    // Mutations
    createTemplate,
    createTemplateLoading,
    
    // Refetch functions
    refetchTemplate
  };
} 