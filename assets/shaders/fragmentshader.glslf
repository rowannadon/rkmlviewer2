#version 330
out vec4 frag_color;
in vec2 Texcoord;
in vec3 Normal;
in vec3 FragPos;
uniform vec3 lightPos;
uniform sampler2D ourTexture;
uniform sampler2D ourTexture2;
uniform sampler2D specMap;
uniform vec3 objectColor;
uniform vec3 lightColor;
uniform vec3 viewPos;
uniform vec3 atmoColor;
uniform vec3 atmoColor2;
uniform float ambientStrength;
vec3 lightStrength;

void main() {
    vec3 origin = vec3(0.0,0.0,0.0);
    float dist = distance(origin, viewPos);
    float cutoff = min(1.0, clamp(dist-1.5, 0.0, 1.0));

    float specularStrength = 1.1;
    float atmosphereStrength = 1.2;
    float atmosphereStrength2 = 1.4;
    vec3 ambient = ambientStrength * lightColor;

    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * lightColor;

    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);

    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 16);
    vec3 specular = specularStrength * spec * lightColor * vec3(texture(specMap, Texcoord));

    float atmo = pow(1 - dot(norm, viewDir), 8);
    float atmo2 = pow(1 - dot(norm, viewDir), 1);
    vec3 atmosphere = atmosphereStrength * (diffuse+(ambient*0.6)) * atmo * atmoColor * cutoff;
    vec3 atmosphere2 = atmosphereStrength2 * (diffuse+(ambient*0.1)) * atmo2 * atmoColor2 * clamp(cutoff+0.6, 0.0, 1.0);

    lightStrength = vec3(12.0, 12.0, 24.0);
    float lightIntensity = max(0.0, 0.8-ambientStrength);

    vec3 lights = pow(vec3(texture(ourTexture2, Texcoord)) * pow((1-diff), 3), lightStrength) * lightIntensity * cutoff;

    vec3 result = (ambient + diffuse*clamp(1-ambientStrength, 0.0, 1.0) + specular) * objectColor;

    vec4 front = texture(ourTexture, Texcoord);

    frag_color = front * vec4(result, 1.0) + vec4(atmosphere2+atmosphere+lights, 1.0);
}