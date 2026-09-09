export class RestConstants {
    public readonly API_URL = 'http://localhost:8080/cdv-api/';

    public getApiURL(): string {
        return this.API_URL;
    }
}