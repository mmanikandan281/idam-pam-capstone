import React, { useEffect, useState } from 'react';
import { Users, Key, Shield, Activity, TrendingUp } from 'lucide-react';
import { auditApi, userApi, secretApi } from '../services/api';

interface Stats {
  totalUsers: number;
  totalSecrets: number;
  recentLogins: number;
  totalAuditLogs: number;
}

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<Stats>({
    totalUsers: 0,
    totalSecrets: 0,
    recentLogins: 0,
    totalAuditLogs: 0,
  });
  const [recentActivity, setRecentActivity] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      const [users, secrets, auditLogs] = await Promise.all([
        userApi.getUsers(),
        secretApi.getSecrets(),
        auditApi.getAuditLogs(10, 0),
      ]);

      // Calculate recent logins (last 24 hours)
      const oneDayAgo = new Date();
      oneDayAgo.setDate(oneDayAgo.getDate() - 1);
      
      const recentLogins = auditLogs.filter((log: any) => 
        log.action.includes('login.success') && 
        new Date(log.created_at) > oneDayAgo
      ).length;

      setStats({
        totalUsers: users.length,
        totalSecrets: secrets.length,
        recentLogins,
        totalAuditLogs: auditLogs.length,
      });

      setRecentActivity(auditLogs.slice(0, 5));
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const statCards = [
    {
      title: 'Total Users',
      value: stats.totalUsers,
      icon: Users,
      color: 'bg-blue-500',
      change: '+2.5%',
    },
    {
      title: 'Stored Secrets',
      value: stats.totalSecrets,
      icon: Key,
      color: 'bg-green-500',
      change: '+12%',
    },
    {
      title: 'Recent Logins',
      value: stats.recentLogins,
      icon: Activity,
      color: 'bg-purple-500',
      change: '+5.2%',
    },
    {
      title: 'Audit Logs',
      value: stats.totalAuditLogs,
      icon: Shield,
      color: 'bg-orange-500',
      change: '+8.1%',
    },
  ];

  const getActionIcon = (action: string) => {
    if (action.includes('login')) return '🔐';
    if (action.includes('secret')) return '🔑';
    if (action.includes('user')) return '👤';
    return '📝';
  };

  const getActionColor = (action: string) => {
    if (action.includes('login.success')) return 'text-green-600 bg-green-50';
    if (action.includes('login.failed')) return 'text-red-600 bg-red-50';
    if (action.includes('create')) return 'text-blue-600 bg-blue-50';
    if (action.includes('delete')) return 'text-red-600 bg-red-50';
    return 'text-gray-600 bg-gray-50';
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl p-6 text-white">
        <h1 className="text-3xl font-bold mb-2">Welcome to IDAM-PAM Platform</h1>
        <p className="text-blue-100">
          Manage identities, access controls, and privileged credentials securely.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => {
          const Icon = stat.icon;
          return (
            <div key={index} className="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow duration-200">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600 mb-2">{stat.title}</p>
                  <p className="text-3xl font-bold text-gray-900">{stat.value}</p>
                </div>
                <div className={`${stat.color} p-3 rounded-lg`}>
                  <Icon className="h-6 w-6 text-white" />
                </div>
              </div>
              <div className="mt-4 flex items-center">
                <TrendingUp className="h-4 w-4 text-green-500 mr-1" />
                <span className="text-sm font-medium text-green-600">{stat.change}</span>
                <span className="text-sm text-gray-500 ml-2">vs last month</span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-xl shadow-sm">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Recent Activity</h2>
        </div>
        <div className="p-6">
          {recentActivity.length === 0 ? (
            <div className="text-center py-8">
              <Activity className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No recent activity</h3>
              <p className="mt-1 text-sm text-gray-500">
                Activity will appear here as users interact with the system.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {recentActivity.map((activity, index) => (
                <div key={index} className="flex items-center space-x-4 p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors duration-200">
                  <div className="flex-shrink-0">
                    <span className="text-2xl">{getActionIcon(activity.action)}</span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getActionColor(activity.action)}`}>
                        {activity.action}
                      </span>
                      {activity.username && (
                        <span className="text-sm text-gray-600">by {activity.username}</span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 mt-1">
                      {activity.resource} • {new Date(activity.created_at).toLocaleDateString()} at {new Date(activity.created_at).toLocaleTimeString()}
                    </p>
                  </div>
                  <div className="flex-shrink-0">
                    <span className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-200 text-gray-800">
                      {activity.ip_address}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-xl shadow-sm">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Quick Actions</h2>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors duration-200 group">
              <Users className="h-8 w-8 text-gray-400 group-hover:text-blue-500 mx-auto mb-2" />
              <p className="text-sm font-medium text-gray-900 group-hover:text-blue-600">Add New User</p>
              <p className="text-xs text-gray-500 mt-1">Create a new user account</p>
            </button>
            
            <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-green-500 hover:bg-green-50 transition-colors duration-200 group">
              <Key className="h-8 w-8 text-gray-400 group-hover:text-green-500 mx-auto mb-2" />
              <p className="text-sm font-medium text-gray-900 group-hover:text-green-600">Store Secret</p>
              <p className="text-xs text-gray-500 mt-1">Add a new credential to vault</p>
            </button>
            
            <button className="p-4 border-2 border-dashed border-gray-300 rounded-lg hover:border-purple-500 hover:bg-purple-50 transition-colors duration-200 group">
              <Shield className="h-8 w-8 text-gray-400 group-hover:text-purple-500 mx-auto mb-2" />
              <p className="text-sm font-medium text-gray-900 group-hover:text-purple-600">View Audit Logs</p>
              <p className="text-xs text-gray-500 mt-1">Review system activity</p>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;