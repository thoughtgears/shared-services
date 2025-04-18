import React from 'react';
import {auth} from '../firebase';
import {useNavigate} from 'react-router-dom';

const Dashboard: React.FC = () => {
    const navigate = useNavigate();
    const user = auth.currentUser;

    const handleLogout = async () => {
        try {
            await auth.signOut();
            navigate('/login');
        } catch (error) {
            console.error('Error logging out:', error);
        }
    };

    return (
        <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
            <div className="container mx-auto px-4 py-8">
                <div className="flex justify-between items-center mb-8">
                    <h1 className="text-2xl font-bold text-gray-800 dark:text-white">Dashboard</h1>
                    <button
                        onClick={handleLogout}
                        className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md"
                    >
                        Logout
                    </button>
                </div>

                <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
                    <h2 className="text-xl font-semibold text-gray-800 dark:text-white mb-4">
                        Welcome, {user?.email}!
                    </h2>
                    <p className="text-gray-600 dark:text-gray-300">
                        You've successfully logged in to the dashboard.
                    </p>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;