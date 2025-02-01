use eframe::egui;
use rfd::FileDialog;
use std::{
    env,
    fs::{File, OpenOptions},
    io::{BufReader, BufWriter, Write},
    path::PathBuf,
};

fn main() -> eframe::Result<()> {
    let options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default()
            .with_inner_size([600.0, 600.0]),
        ..Default::default()
    };
    
    eframe::run_native(
        "CSV Sorter",
        options,
        Box::new(|_cc| Box::<CsvSorterApp>::default()),
    )
}

#[derive(Default)]
struct CsvSorterApp {
    last_message: String,
    output_path: Option<PathBuf>,
}

impl eframe::App for CsvSorterApp {
    fn update(&mut self, ctx: &egui::Context, _frame: &mut eframe::Frame) {
        egui::CentralPanel::default().show(ctx, |ui| {
            ui.vertical_centered(|ui| {
                ui.add_space(200.0); // Add space at the top
                
                let button = egui::Button::new("Select CSV File")
                    .min_size(egui::vec2(200.0, 50.0));
                
                if ui.add(button).clicked() {
                    if let Some(path) = FileDialog::new()
                        .add_filter("CSV", &["csv"])
                        .set_directory(env::current_dir().unwrap_or_default())
                        .pick_file() 
                    {
                        self.process_csv(path);
                    }
                }
                
                if !self.last_message.is_empty() {
                    ui.add_space(20.0);
                    ui.label(&self.last_message);
                    
                    if let Some(path) = &self.output_path {
                        ui.add_space(10.0);
                        ui.label(format!("Output file: {}", path.display()));
                    }
                }
            });
        });
    }
}

impl CsvSorterApp {
    fn process_csv(&mut self, path: PathBuf) {
        match self.handle_csv_processing(path) {
            Ok(msg) => self.last_message = msg,
            Err(e) => {
                self.last_message = format!("Error: {}", e);
                self.output_path = None;
            },
        }
    }

    fn handle_csv_processing(&mut self, path: PathBuf) -> Result<String, Box<dyn std::error::Error>> {
        let file = File::open(&path)?;
        let reader = BufReader::new(file);
        let mut csv_reader = csv::Reader::from_reader(reader);
        
        let headers = csv_reader.headers()?.clone();
        let tier_idx = headers.iter().position(|h| h == "Tier")
            .ok_or("Tier column not found")?;
        let status_idx = headers.iter().position(|h| h == "Patron Status")
            .ok_or("Patron Status column not found")?;
        let name_idx = headers.iter().position(|h| h == "Name")
            .ok_or("Name column not found")?;

        let mut records: Vec<csv::StringRecord> = csv_reader
            .records()
            .filter_map(Result::ok)
            .filter(|record| {
                record.get(status_idx)
                    .map(|status| status.trim().eq_ignore_ascii_case("Active Patron"))
                    .unwrap_or(false)
            })
            .collect();

        records.sort_by(|a, b| {
            a.get(tier_idx).unwrap_or("")
                .cmp(&b.get(tier_idx).unwrap_or(""))
        });

        let mut output = String::new();
        let mut prev_tier = String::new();

        for record in records {
            let current_tier = record.get(tier_idx).unwrap_or("");
            if current_tier != prev_tier {
                output.push_str("---------------------------\n");
                prev_tier = current_tier.to_string();
            }
            if let Some(name) = record.get(name_idx) {
                output.push_str(&format!("{}\n", name));
            }
        }

        let output_path = path.with_file_name(
            format!("{}_sort.txt", 
                path.file_stem().unwrap().to_string_lossy())
        );

        let output_file = OpenOptions::new()
            .write(true)
            .create(true)
            .truncate(true)
            .open(&output_path)?;
            
        let mut writer = BufWriter::new(output_file);
        writer.write_all(output.as_bytes())?;

        self.output_path = Some(output_path);
        Ok(format!("Export completed successfully"))
    }
}
