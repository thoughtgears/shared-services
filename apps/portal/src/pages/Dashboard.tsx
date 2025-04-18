import React, { useState, useEffect } from 'react';
import { auth } from '../lib/firebase';
import { useNavigate } from 'react-router-dom';

interface UserData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  firebase_id: string;
  address: {
    building_number: string;
    street: string;
    city: string;
    postcode: string;
    country: string;
  };
}

interface AddressFormData {
  building_number: string;
  street: string;
  city: string;
  postcode: string;
  country: string;
}

interface Document {
  id: string;
  user_id: string;
  name: string;
  size: number;
  type: string;
  content_type: string;
  path: string;
  bucket: string;
  created_at: string;
  updated_at: string;
}

const API_ENDPOINT = import.meta.env.VITE_API_GATEWAY_ENDPOINT || '';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const user = auth.currentUser;
  const [userData, setUserData] = useState<UserData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [isEditingAddress, setIsEditingAddress] = useState<boolean>(false);
  const [addressForm, setAddressForm] = useState<AddressFormData>({
    building_number: '',
    street: '',
    city: '',
    postcode: '',
    country: ''
  });
  const [updateSuccess, setUpdateSuccess] = useState<boolean>(false);
  const [updateLoading, setUpdateLoading] = useState<boolean>(false);
  
  // Document state
  const [documents, setDocuments] = useState<Document[]>([]);
  const [documentsLoading, setDocumentsLoading] = useState<boolean>(false);
  const [documentsError, setDocumentsError] = useState<string>('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [documentType, setDocumentType] = useState<string>('passport');
  const [uploadProgress, setUploadProgress] = useState<boolean>(false);
  const [documentSuccess, setDocumentSuccess] = useState<string>('');

  useEffect(() => {
    const fetchUserData = async () => {
      if (!user) {
        setLoading(false);
        return;
      }

      try {
        // Get the Firebase ID token for authentication
        const idToken = await user.getIdToken();
        
        // Fetch user data from your API using the Firebase ID
        const response = await fetch(`${API_ENDPOINT}/v1/users/${user.uid}`, {
          headers: {
            'Authorization': `Bearer ${idToken}`
          }
        });

        if (!response.ok) {
          throw new Error(`Error fetching user data: ${response.status}`);
        }

        const responseData = await response.json();
        setUserData(responseData.data);
        
        // Initialize the address form with current data
        if (responseData.data?.address) {
          setAddressForm(responseData.data.address);
        }
        
        // After getting user data, fetch their documents
        fetchUserDocuments(responseData.data.id, idToken);
      } catch (err) {
        console.error('Error fetching user data:', err);
        setError('Failed to load user data. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchUserData();
  }, [user]);

  const fetchUserDocuments = async (userId: string, idToken: string) => {
    setDocumentsLoading(true);
    setDocumentsError('');

    try {
      const response = await fetch(`${API_ENDPOINT}/v1/documents?user_id=${userId}`, {
        headers: {
          'Authorization': `Bearer ${idToken}`
        }
      });

      if (!response.ok) {
        throw new Error(`Error fetching documents: ${response.status}`);
      }

      const responseData = await response.json();
      setDocuments(responseData.data || []);
    } catch (err) {
      console.error('Error fetching user documents:', err);
      setDocumentsError('Failed to load documents. Please try again later.');
    } finally {
      setDocumentsLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await auth.signOut();
      navigate('/login');
    } catch (error) {
      console.error('Error logging out:', error);
    }
  };

  const handleAddressChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setAddressForm(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const startEditingAddress = () => {
    setIsEditingAddress(true);
    setUpdateSuccess(false);
  };

  const cancelEditingAddress = () => {
    if (userData?.address) {
      setAddressForm(userData.address);
    }
    setIsEditingAddress(false);
    setUpdateSuccess(false);
  };

  const saveAddressChanges = async () => {
    if (!userData || !user) return;
    
    setUpdateLoading(true);
    try {
      const idToken = await user.getIdToken();
      
      // Create the update payload with only the address changes
      const updatePayload = {
        address: addressForm
      };
      
      const response = await fetch(`${API_ENDPOINT}/v1/users/${userData.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${idToken}`
        },
        body: JSON.stringify(updatePayload)
      });

      if (!response.ok) {
        throw new Error(`Error updating address: ${response.status}`);
      }

      const updatedData = await response.json();
      setUserData(updatedData.data);
      setIsEditingAddress(false);
      setUpdateSuccess(true);
      
      // Show success message temporarily
      setTimeout(() => {
        setUpdateSuccess(false);
      }, 3000);
    } catch (err) {
      console.error('Error updating address:', err);
      setError('Failed to update address. Please try again.');
    } finally {
      setUpdateLoading(false);
    }
  };

  // Document handlers
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleDocumentTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setDocumentType(e.target.value);
  };

  const uploadDocument = async () => {
    if (!selectedFile || !userData) return;
    
    setUploadProgress(true);
    setDocumentSuccess('');
    setDocumentsError('');

    try {
      const idToken = await user?.getIdToken();
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('user_id', userData.id);
      formData.append('document_type', documentType);
      
      const response = await fetch(`${API_ENDPOINT}/v1/documents`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${idToken}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`Error uploading document: ${response.status}`);
      }

      // Reset form
      setSelectedFile(null);
      const fileInput = document.getElementById('file-input') as HTMLInputElement;
      if (fileInput) fileInput.value = '';

      // Show success message
      setDocumentSuccess('Document uploaded successfully');
      setTimeout(() => setDocumentSuccess(''), 3000);
      
      // Refresh document list
      fetchUserDocuments(userData.id, idToken as string || '');
    } catch (err) {
      console.error('Error uploading document:', err);
      setDocumentsError('Failed to upload document. Please try again.');
    } finally {
      setUploadProgress(false);
    }
  };

  const deleteDocument = async (documentId: string) => {
    if (!userData || !user) return;
    
    if (!confirm('Are you sure you want to delete this document?')) return;
    
    try {
      const idToken = await user.getIdToken();
      const response = await fetch(`${API_ENDPOINT}/v1/documents/${documentId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${idToken}`
        }
      });

      if (!response.ok) {
        throw new Error(`Error deleting document: ${response.status}`);
      }
      
      // Show success message
      setDocumentSuccess('Document deleted successfully');
      setTimeout(() => setDocumentSuccess(''), 3000);
      
      // Remove from list
      setDocuments(documents.filter(doc => doc.id !== documentId));
    } catch (err) {
      console.error('Error deleting document:', err);
      setDocumentsError('Failed to delete document. Please try again.');
    }
  };

  // Format file size for display
  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + ' bytes';
    else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    else return (bytes / 1048576).toFixed(1) + ' MB';
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
        <div className="animate-pulse flex flex-col items-center">
          <div className="w-20 h-20 bg-blue-400 dark:bg-blue-600 rounded-full mb-4"></div>
          <div className="h-4 w-36 bg-gray-300 dark:bg-gray-700 rounded mb-3"></div>
          <div className="h-3 w-24 bg-gray-200 dark:bg-gray-800 rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <nav className="bg-white dark:bg-gray-800 shadow-md">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-xl font-semibold text-gray-800 dark:text-white">User Portal</h1>
          <button
            onClick={handleLogout}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
          >
            Logout
          </button>
        </div>
      </nav>
      
      <div className="container mx-auto px-4 py-8">
        {error ? (
          <div className="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 mb-6" role="alert">
            <p>{error}</p>
          </div>
        ) : null}

        {updateSuccess && (
          <div className="bg-green-100 border-l-4 border-green-500 text-green-700 p-4 mb-6 animate-fadeIn" role="alert">
            <p>Address updated successfully!</p>
          </div>
        )}

        {documentSuccess && (
          <div className="bg-green-100 border-l-4 border-green-500 text-green-700 p-4 mb-6 animate-fadeIn" role="alert">
            <p>{documentSuccess}</p>
          </div>
        )}

        {documentsError && (
          <div className="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 mb-6" role="alert">
            <p>{documentsError}</p>
          </div>
        )}

        {userData ? (
          <div className="grid grid-cols-1 gap-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Profile Summary Card */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden">
                <div className="bg-blue-600 dark:bg-blue-800 p-6 flex flex-col items-center">
                  <div className="w-24 h-24 bg-white dark:bg-gray-200 rounded-full flex items-center justify-center text-blue-600 text-3xl font-bold mb-4">
                    {userData.first_name.charAt(0)}{userData.last_name.charAt(0)}
                  </div>
                  <h2 className="text-2xl font-bold text-white">
                    {userData.first_name} {userData.last_name}
                  </h2>
                  <p className="text-blue-100 mt-1">{userData.email}</p>
                </div>
                <div className="p-6">
                  <div className="mb-4">
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Contact Information</h3>
                    <p className="text-gray-800 dark:text-gray-200 flex items-center">
                      <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
                      </svg>
                      {userData.email}
                    </p>
                    <p className="text-gray-800 dark:text-gray-200 flex items-center mt-2">
                      <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
                      </svg>
                      {userData.phone || 'No phone number provided'}
                    </p>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">Account Details</h3>
                    <p className="text-gray-800 dark:text-gray-200 flex items-center">
                      <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 6H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2h-5m-4 0V5a2 2 0 114 0v1m-4 0a2 2 0 104 0m-5 8a2 2 0 100-4 2 2 0 000 4zm0 0c1.306 0 2.417.835 2.83 2M9 14a3.001 3.001 0 00-2.83 2M15 11h3m-3 4h2"></path>
                      </svg>
                      User ID: {userData.id.substring(0, 8)}...
                    </p>
                    <p className="text-gray-800 dark:text-gray-200 flex items-center mt-2">
                      <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4"></path>
                      </svg>
                      Firebase ID: {userData.firebase_id.substring(0, 8)}...
                    </p>
                  </div>
                </div>
              </div>

              {/* Address Card */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden col-span-1 md:col-span-2">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
                  <h2 className="text-xl font-semibold text-gray-800 dark:text-white">Address</h2>
                  {!isEditingAddress ? (
                    <button 
                      onClick={startEditingAddress}
                      className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors"
                    >
                      Edit Address
                    </button>
                  ) : (
                    <div className="space-x-2">
                      <button 
                        onClick={cancelEditingAddress}
                        className="px-4 py-2 bg-gray-300 hover:bg-gray-400 text-gray-800 rounded-md transition-colors"
                      >
                        Cancel
                      </button>
                      <button 
                        onClick={saveAddressChanges}
                        disabled={updateLoading}
                        className={`px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-colors ${updateLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
                      >
                        {updateLoading ? 'Saving...' : 'Save Changes'}
                      </button>
                    </div>
                  )}
                </div>

                <div className="p-6">
                  {isEditingAddress ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Building Number</label>
                        <input
                          type="text"
                          name="building_number"
                          value={addressForm.building_number}
                          onChange={handleAddressChange}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Street</label>
                        <input
                          type="text"
                          name="street"
                          value={addressForm.street}
                          onChange={handleAddressChange}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">City</label>
                        <input
                          type="text"
                          name="city"
                          value={addressForm.city}
                          onChange={handleAddressChange}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Postal Code</label>
                        <input
                          type="text"
                          name="postcode"
                          value={addressForm.postcode}
                          onChange={handleAddressChange}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Country</label>
                        <input
                          type="text"
                          name="country"
                          value={addressForm.country}
                          onChange={handleAddressChange}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                        />
                      </div>
                    </div>
                  ) : (
                    <div className="bg-gray-50 dark:bg-gray-700 p-5 rounded-lg">
                      {userData.address ? (
                        <div className="flex">
                          <svg className="w-6 h-6 text-blue-500 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path>
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path>
                          </svg>
                          <div>
                            <p className="text-lg font-medium text-gray-800 dark:text-white">
                              {userData.address.building_number} {userData.address.street}
                            </p>
                            <p className="text-gray-600 dark:text-gray-300">
                              {userData.address.city}, {userData.address.postcode}
                            </p>
                            <p className="text-gray-600 dark:text-gray-300">
                              {userData.address.country}
                            </p>
                          </div>
                        </div>
                      ) : (
                        <p className="text-gray-600 dark:text-gray-300">No address information available.</p>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Documents Section */}
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden">
              <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                <h2 className="text-xl font-semibold text-gray-800 dark:text-white">Documents</h2>
              </div>

              {/* Upload Document Form */}
              <div className="p-6 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700">
                <h3 className="text-lg font-medium text-gray-800 dark:text-white mb-4">Upload New Document</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="col-span-1">
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Document Type</label>
                    <select
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      value={documentType}
                      onChange={handleDocumentTypeChange}
                    >
                      <option value="passport">Passport</option>
                      <option value="id_card">ID Card</option>
                      <option value="driver_licence">Driver's License</option>
                      <option value="other">Other</option>
                    </select>
                  </div>
                  <div className="col-span-1">
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">File</label>
                    <input
                      id="file-input"
                      type="file"
                      onChange={handleFileChange}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                  </div>
                  <div className="col-span-1 flex items-end">
                    <button
                      onClick={uploadDocument}
                      disabled={!selectedFile || uploadProgress}
                      className={`px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors w-full ${(!selectedFile || uploadProgress) ? 'opacity-70 cursor-not-allowed' : ''}`}
                    >
                      {uploadProgress ? 'Uploading...' : 'Upload Document'}
                    </button>
                  </div>
                </div>
              </div>

              {/* Document List */}
              <div className="p-6">
                <h3 className="text-lg font-medium text-gray-800 dark:text-white mb-4">Your Documents</h3>
                
                {documentsLoading ? (
                  <div className="flex justify-center py-8">
                    <div className="animate-pulse flex flex-col items-center">
                      <div className="h-4 w-36 bg-gray-300 dark:bg-gray-700 rounded mb-3"></div>
                      <div className="h-3 w-24 bg-gray-200 dark:bg-gray-800 rounded"></div>
                    </div>
                  </div>
                ) : documents.length === 0 ? (
                  <div className="text-center py-8 bg-gray-50 dark:bg-gray-700 rounded-lg">
                    <svg className="w-12 h-12 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                    </svg>
                    <p className="text-gray-600 dark:text-gray-300 text-lg">No documents found. Upload your first document above.</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                      <thead className="bg-gray-50 dark:bg-gray-700">
                        <tr>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                            Type
                          </th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                            Name
                          </th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                            Size
                          </th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                            Uploaded
                          </th>
                          <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                            Actions
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                        {documents.map((doc) => (
                          <tr key={doc.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100">
                                {doc.type}
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
                              {doc.name}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {formatFileSize(doc.size)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                              {new Date(doc.created_at).toLocaleDateString()}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                              <button
                                onClick={() => deleteDocument(doc.id)}
                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                              >
                                Delete
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md text-center">
            <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
            </svg>
            <p className="text-gray-600 dark:text-gray-300 text-lg">
              No user data available. Please make sure your account is properly set up.
            </p>
            <button
              onClick={handleLogout}
              className="mt-4 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors"
            >
              Return to Login
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard;