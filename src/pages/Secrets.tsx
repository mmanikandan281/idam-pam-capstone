import React, { useEffect, useState } from 'react';
import { Key, Plus, Eye, EyeOff, Trash2, Copy } from 'lucide-react';
import { secretApi } from '../services/api';
import toast from 'react-hot-toast';

interface Secret {
  id: string;
  name: string;
  description: string;
  created_by: string;
  created_by_username: string;
  created_at: string;
  updated_at: string;
  data?: string;
}

const Secrets: React.FC = () => {
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [createData, setCreateData] = useState({
    name: '',
    description: '',
    data: '',
  });
  const [visibleSecrets, setVisibleSecrets] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadSecrets();
  }, []);

  const loadSecrets = async () => {
    try {
      const data = await secretApi.getSecrets();
      setSecrets(data);
    } catch (error) {
      toast.error('Failed to load secrets');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateSecret = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await secretApi.createSecret(createData);
      toast.success('Secret created successfully');
      setIsCreateModalOpen(false);
      setCreateData({ name: '', description: '', data: '' });
      loadSecrets();
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  const toggleSecretVisibility = async (secretId: string) => {
    if (visibleSecrets.has(secretId)) {
      // Hide secret
      setVisibleSecrets(prev => {
        const newSet = new Set(prev);
        newSet.delete(secretId);
        return newSet;
      });
    } else {
      // Show secret - fetch the actual data
      try {
        const secretData = await secretApi.getSecret(secretId);
        setSecrets(prevSecrets => 
          prevSecrets.map(secret => 
            secret.id === secretId 
              ? { ...secret, data: secretData.data }
              : secret
          )
        );
        setVisibleSecrets(prev => new Set(prev).add(secretId));
      } catch (error) {
        toast.error('Failed to decrypt secret');
      }
    }
  };

  const copyToClipboard = async (data: string) => {
    try {
      await navigator.clipboard.writeText(data);
      toast.success('Secret copied to clipboard');
    } catch (error) {
      toast.error('Failed to copy to clipboard');
    }
  };

  const deleteSecret = async (secretId: string) => {
    if (!confirm('Are you sure you want to delete this secret?')) return;
    
    try {
      await secretApi.deleteSecret(secretId);
      toast.success('Secret deleted successfully');
      loadSecrets();
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Secret Vault</h1>
          <p className="text-gray-600 mt-1">Securely store and manage privileged credentials</p>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition duration-200"
        >
          <Plus className="h-4 w-4 mr-2" />
          Store Secret
        </button>
      </div>

      {/* Secrets Grid */}
      {secrets.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <Key className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No secrets stored</h3>
          <p className="mt-1 text-sm text-gray-500">
            Get started by storing your first credential.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {secrets.map((secret) => (
            <div key={secret.id} className="bg-white rounded-xl shadow-sm hover:shadow-md transition-shadow duration-200 overflow-hidden">
              <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <Key className="h-8 w-8 text-blue-600" />
                    </div>
                    <div className="ml-3">
                      <h3 className="text-lg font-medium text-gray-900">{secret.name}</h3>
                    </div>
                  </div>
                  <button
                    onClick={() => deleteSecret(secret.id)}
                    className="text-red-400 hover:text-red-600 transition-colors duration-200"
                  >
                    <Trash2 className="h-5 w-5" />
                  </button>
                </div>

                <p className="text-gray-600 text-sm mb-4">{secret.description}</p>

                {/* Secret Data */}
                <div className="bg-gray-50 rounded-lg p-3 mb-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      {visibleSecrets.has(secret.id) ? (
                        <div className="font-mono text-sm text-gray-900 break-all">
                          {secret.data || '••••••••'}
                        </div>
                      ) : (
                        <div className="text-sm text-gray-500">••••••••••••••••</div>
                      )}
                    </div>
                    <div className="flex space-x-2 ml-2">
                      <button
                        onClick={() => toggleSecretVisibility(secret.id)}
                        className="text-gray-400 hover:text-gray-600 transition-colors duration-200"
                      >
                        {visibleSecrets.has(secret.id) ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                      {visibleSecrets.has(secret.id) && secret.data && (
                        <button
                          onClick={() => copyToClipboard(secret.data!)}
                          className="text-gray-400 hover:text-gray-600 transition-colors duration-200"
                        >
                          <Copy className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </div>
                </div>

                {/* Metadata */}
                <div className="text-xs text-gray-500 space-y-1">
                  <div>Created by: {secret.created_by_username}</div>
                  <div>Created: {new Date(secret.created_at).toLocaleDateString()}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Secret Modal */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <form onSubmit={handleCreateSecret}>
                <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                  <div className="sm:flex sm:items-start">
                    <div className="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10">
                      <Key className="h-6 w-6 text-blue-600" />
                    </div>
                    <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
                      <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                        Store New Secret
                      </h3>
                      <div className="space-y-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Name</label>
                          <input
                            type="text"
                            required
                            value={createData.name}
                            onChange={(e) => setCreateData({...createData, name: e.target.value})}
                            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            placeholder="e.g., Database Password"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Description</label>
                          <input
                            type="text"
                            value={createData.description}
                            onChange={(e) => setCreateData({...createData, description: e.target.value})}
                            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            placeholder="Optional description"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Secret Data</label>
                          <textarea
                            required
                            rows={4}
                            value={createData.data}
                            onChange={(e) => setCreateData({...createData, data: e.target.value})}
                            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                            placeholder="Enter the secret data to encrypt and store"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                  <button
                    type="submit"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                  >
                    Store Secret
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsCreateModalOpen(false)}
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Secrets;