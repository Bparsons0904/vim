import { Component } from "solid-js";
import { useAuth } from "@context/AuthContext";
import { Button } from "@components/common/ui/Button/Button";

const Profile: Component = () => {
  const { user, logout } = useAuth();

  return (
    <div style={{ padding: "2rem", "max-width": "600px", margin: "0 auto" }}>
      <h1>Profile</h1>
      <div style={{ margin: "2rem 0" }}>
        <h2>User Information</h2>
        <div style={{ margin: "1rem 0" }}>
          <p>
            <strong>Username:</strong> {user?.login || "Not available"}
          </p>
          <p>
            <strong>Name:</strong> {user?.firstName || ""}{" "}
            {user?.lastName || ""}
          </p>
          <p>
            <strong>User ID:</strong> {user?.id || "Not available"}
          </p>
          <p>
            <strong>Member Since:</strong>{" "}
            {user?.createdAt
              ? new Date(user.createdAt).toLocaleDateString()
              : "Not available"}
          </p>
        </div>
      </div>
      <div>
        <Button variant="primary" onClick={logout}>
          Sign Out
        </Button>
      </div>
    </div>
  );
};

export default Profile;

