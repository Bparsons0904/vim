import { Component } from "solid-js";
import { useAuth } from "@context/AuthContext";

const Profile: Component = () => {
  const { user, logout } = useAuth();

  return (
    <div>
      <h1>Profile</h1>
      <div>
        <h2>User Information</h2>
        <p><strong>Login:</strong> {user?.login || "Not available"}</p>
        <p><strong>Name:</strong> {user?.firstName || ""} {user?.lastName || ""}</p>
        <p><strong>ID:</strong> {user?.id || "Not available"}</p>
        <p><strong>Created:</strong> {user?.createdAt ? new Date(user.createdAt).toLocaleDateString() : "Not available"}</p>
      </div>
      <div>
        <button onClick={logout}>Logout</button>
      </div>
    </div>
  );
};

export default Profile;