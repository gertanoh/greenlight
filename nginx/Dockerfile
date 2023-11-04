# Use the official NGINX image as a parent image
FROM nginx:latest


# Copy the configuration file from the current directory to the container
COPY nginx.conf /etc/nginx/conf.d/default.conf
# The port that your container listens on.
EXPOSE 80

# Start NGINX when the container launches
CMD ["nginx", "-g", "daemon off;"]
