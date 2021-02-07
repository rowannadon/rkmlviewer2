#version 330
layout (location = 0) in vec3 vp;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 texcoord;
out vec2 Texcoord;
out vec3 Normal;
out vec3 FragPos;

uniform mat4 model;
uniform mat4 camera;
uniform mat4 projection;

void main() {
    gl_Position = projection * camera * model * vec4(vp, 1.0);
    FragPos = vec3(model * vec4(vp, 1.0));
    Texcoord = texcoord;
    Normal = mat3(transpose(inverse(model))) * aNormal;
}