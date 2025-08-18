import { useState } from "react";
import { useAuth } from "../../core/auth/AuthContext";

// Profile Component
export const ProfilePage = ({ onBack }: { onBack: () => void }) => {
  const { user, apiService } = useAuth();
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [displayName, setDisplayName] = useState(user?.displayName || '');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    try {
      const updateData: any = { displayName };
      
      if (newPassword) {
        if (newPassword !== confirmPassword) {
          throw new Error('New passwords do not match');
        }
        if (newPassword.length < 6) {
          throw new Error('Password must be at least 6 characters long');
        }
        updateData.currentPassword = currentPassword;
        updateData.newPassword = newPassword;
      }

      await apiService.updateProfile(updateData);
      
      setMessage('Profile updated successfully!');
      setDisplayName(updateData?.displayName || '');

      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error: any) {
      setMessage(error.message);
    }
    
    setLoading(false);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 p-4">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center gap-4 mb-8">
          <button
            onClick={onBack}
            className="px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition duration-200"
          >
            ← Back
          </button>
          <h1 className="text-3xl font-bold text-white">Edit Profile</h1>
        </div>

        <div className="bg-white/10 backdrop-blur-md rounded-xl p-8 border border-white/20">
          <form onSubmit={handleUpdateProfile} className="space-y-6">
            {message && (
              <div className={`border rounded-lg p-3 text-sm ${
                message.includes('success') 
                  ? 'bg-green-500/20 border-green-500/30 text-green-100' 
                  : 'bg-red-500/20 border-red-500/30 text-red-100'
              }`}>
                {message}
              </div>
            )}

            <div>
              <label className="block text-gray-300 text-sm font-medium mb-2">Email</label>
              <input
                type="email"
                value={user?.email || ''}
                disabled
                className="w-full px-4 py-3 bg-gray-700/50 border border-gray-600 rounded-lg text-gray-400 cursor-not-allowed"
              />
              <p className="text-xs text-gray-400 mt-1">Email cannot be changed</p>
            </div>

            <div>
              <label className="block text-gray-300 text-sm font-medium mb-2">Display Name</label>
              <input
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>

            <div className="border-t border-white/20 pt-6">
              <h3 className="text-lg font-medium text-white mb-4">Change Password</h3>
              <p className="text-sm text-gray-400 mb-4">Leave blank to keep current password</p>

              <div className="space-y-4">
                <div>
                  <label className="block text-gray-300 text-sm font-medium mb-2">Current Password</label>
                  <input
                    type="password"
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Enter current password"
                  />
                </div>

                <div>
                  <label className="block text-gray-300 text-sm font-medium mb-2">New Password</label>
                  <input
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Enter new password"
                  />
                </div>

                <div>
                  <label className="block text-gray-300 text-sm font-medium mb-2">Confirm New Password</label>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Confirm new password"
                  />
                </div>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-lg font-medium hover:from-blue-700 hover:to-purple-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Updating...' : 'Update Profile'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
