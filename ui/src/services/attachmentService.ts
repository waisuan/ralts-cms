import { apiClient, ApiResponse, ApiError } from '../utils/api';

export class AttachmentService {
  private static readonly BASE_PATH = '/api/v1/machines';

  /**
   * Upload an attachment for a machine
   */
  static async uploadMachineAttachment(
    machineSerialNumber: string,
    file: File
  ): Promise<ApiResponse<{ message: string }>> {
    try {
      const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
      const formData = new FormData();
      formData.append('file', file);

      const endpoint = `${this.BASE_PATH}/${encodedSerialNumber}/attachments`;

      console.log('🔧 API: POST attachment upload', { 
        endpoint,
        fileName: file.name,
        fileSize: file.size 
      });

      // Use the centralized API client for consistency
      const response = await apiClient.postFormData<{ message: string }>(endpoint, formData);

      console.log('🔧 API: POST attachment upload success', { fileName: file.name });
      
      // Return consistent response format
      return {
        data: { message: 'Attachment uploaded successfully' }
      };

    } catch (error) {
      console.error('🔧 API: POST attachment upload error', error);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(
        error instanceof Error ? error.message : 'Upload failed',
        500,
        error
      );
    }
  }

  /**
   * Replace an existing attachment for a machine
   */
  static async replaceMachineAttachment(
    machineSerialNumber: string,
    oldAttachmentName: string,
    newFile: File
  ): Promise<ApiResponse<{ message: string }>> {
    try {
      const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
      const encodedAttachmentName = encodeURIComponent(oldAttachmentName);
      const formData = new FormData();
      formData.append('file', newFile);

      const endpoint = `${this.BASE_PATH}/${encodedSerialNumber}/attachments/${encodedAttachmentName}`;

      console.log('🔧 API: PUT attachment replacement', { 
        endpoint,
        oldFileName: oldAttachmentName,
        newFileName: newFile.name,
        fileSize: newFile.size 
      });

      // Use the centralized API client for consistency
      const response = await apiClient.putFormData<{ message: string }>(endpoint, formData);

      console.log('🔧 API: PUT attachment replacement success', { 
        oldFileName: oldAttachmentName,
        newFileName: newFile.name 
      });
      
      // Return consistent response format
      return {
        data: { message: 'Attachment replaced successfully' }
      };

    } catch (error) {
      console.error('🔧 API: PUT attachment replacement error', error);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(
        error instanceof Error ? error.message : 'Replacement failed',
        500,
        error
      );
    }
  }

  /**
   * Download an attachment for a machine
   */
  static async downloadMachineAttachment(
    machineSerialNumber: string,
    attachmentName: string
  ): Promise<Blob> {
    try {
      const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
      const encodedAttachmentName = encodeURIComponent(attachmentName);
      
      const endpoint = `${this.BASE_PATH}/${encodedSerialNumber}/attachments/${encodedAttachmentName}`;

      console.log('🔧 API: GET attachment download', { 
        endpoint,
        machineSerialNumber,
        attachmentName 
      });

      // Use the centralized API client for consistency
      const blob = await apiClient.getBlob(endpoint);

      console.log('🔧 API: GET attachment download success', { 
        attachmentName,
        blobSize: blob.size 
      });
      
      return blob;

    } catch (error) {
      console.error('🔧 API: GET attachment download error', error);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(
        error instanceof Error ? error.message : 'Download failed',
        500,
        error
      );
    }
  }

  /**
   * Delete an attachment for a machine
   */
  static async deleteMachineAttachment(
    machineSerialNumber: string,
    attachmentName: string
  ): Promise<ApiResponse<{ message: string }>> {
    try {
      const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
      const encodedAttachmentName = encodeURIComponent(attachmentName);
      
      const endpoint = `${this.BASE_PATH}/${encodedSerialNumber}/attachments/${encodedAttachmentName}`;

      console.log('🔧 API: DELETE attachment', { 
        endpoint,
        machineSerialNumber,
        attachmentName 
      });

      // Use the centralized API client for consistency
      await apiClient.delete<void>(endpoint);

      console.log('🔧 API: DELETE attachment success', { attachmentName });
      
      // Return consistent response format
      return {
        data: { message: 'Attachment deleted successfully' }
      };

    } catch (error) {
      console.error('🔧 API: DELETE attachment error', error);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(
        error instanceof Error ? error.message : 'Delete failed',
        500,
        error
      );
    }
  }
}
