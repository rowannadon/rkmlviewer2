#version 330
out vec4 frag_color;
in vec2 Texcoord;
in vec3 Normal;
in vec3 FragPos;
uniform vec3 lightPos;
uniform vec3 viewPos;

uniform float ambientStrength;

uniform sampler2D cloudTexture;


void main() {
    vec4 tex = texture(cloudTexture, Texcoord);

    vec3 origin = vec3(0.0,0.0,0.0);
    float dist = distance(origin, viewPos);
    float cutoff = min(1.0, clamp(dist-1.5, 0.0, 1.0));

    float specularStrength = 0.6;

    vec3 ambient = ambientStrength * vec3(1.0,1.0,1.0);

    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * vec3(1.0, 1.0, 1.0);

    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);

    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32);
    vec3 specular = specularStrength * spec * vec3(1.0, 1.0, 1.0) * vec3(texture(cloudTexture, Texcoord));

    vec3 atmo = pow(1 - dot(norm, viewDir), 5) * vec3(1.0, 1.0, 1.0);

    frag_color = vec4((diffuse+ambient+specular), 1.0) * vec4(tex.r, tex.g, tex.b, ((tex.r+tex.g+tex.b)/3)*(1-atmo)*cutoff);
}