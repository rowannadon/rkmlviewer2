#version 330
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aColor;

uniform mat4 model;
uniform mat4 camera;
uniform mat4 projection;

out vec3 ourColor;

void main()
{
    gl_Position = projection * camera * model * vec4(aPos, 1.0);
    ourColor = aColor;
}